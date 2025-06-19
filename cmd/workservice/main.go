package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	pb "rubr/proto/work"
	"strconv"
	"time"
)

type server struct {
	pb.UnimplementedWorkServiceServer
	db *sql.DB
}

func (s *server) GetStudentsByGroupAndDiscipline(ctx context.Context, req *pb.GetStudentsByGroupAndDisciplineRequest) (*pb.GetStudentsByGroupAndDisciplineResponse, error) {
	resp := &pb.GetStudentsByGroupAndDisciplineResponse{
		Students: make([]*pb.GetStudentsByGroupAndDisciplineResponse_Student, 0),
	}

	// Проверяем связь group_id и discipline_id в groups_in_disciplines
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM groups_in_disciplines 
			WHERE group_id = $1 AND discipline_id = $2
		)`, req.GroupId, req.DisciplineId).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка проверки связи group_id и discipline_id: %v", err)
		resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
		return resp, nil
	}
	if !exists {
		resp.Error = fmt.Sprintf("Группа %d не связана с дисциплиной %d", req.GroupId, req.DisciplineId)
		return resp, nil
	}

	// Запрос для получения студентов
	query := `
		SELECT u.id, u.name, u.surname, u.patronymic, u.email
		FROM users u
		JOIN users_in_groups ug ON u.id = ug.user_id
		WHERE ug.group_id = $1 AND u.role = 'student'
	`
	rows, err := s.db.QueryContext(ctx, query, req.GroupId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
		return resp, nil
	}
	defer rows.Close()

	for rows.Next() {
		var student pb.GetStudentsByGroupAndDisciplineResponse_Student
		var patronymic sql.NullString
		if err := rows.Scan(&student.Id, &student.Name, &student.Surname, &patronymic, &student.Email); err != nil {
			log.Printf("Ошибка сканирования строки: %v", err)
			resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
			return resp, nil
		}
		student.Patronymic = patronymic.String
		resp.Students = append(resp.Students, &student)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации по строкам: %v", err)
		resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
		return resp, nil
	}

	return resp, nil
}

func (s *server) UpdateWork(ctx context.Context, req *pb.UpdateWorkRequest) (*pb.UpdateWorkResponse, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE student_works
		SET status = $1
		WHERE id = $2
	`, req.Status, req.WorkId)
	if err != nil {
		log.Printf("Failed to update work %d: %v", req.WorkId, err)
		return &pb.UpdateWorkResponse{Error: err.Error()}, err
	}
	return &pb.UpdateWorkResponse{}, nil
}
func (s *server) GetStudentWorksByTask(ctx context.Context, req *pb.GetStudentWorksByTaskRequest) (*pb.GetStudentWorksByTaskResponse, error) {
	query := `
		SELECT sw.id, u.name, u.surname, u.patronymic, u.email, sw.status, sw.assistant_id,
		       COALESCE(a.name, '') AS assistant_name, COALESCE(a.surname, '') AS assistant_surname, COALESCE(a.patronymic, '') AS assistant_patronymic,
		       sw.student_id
		FROM student_works sw
		JOIN users u ON sw.student_id = u.id
		LEFT JOIN users a ON sw.assistant_id = a.id
		WHERE sw.task_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		log.Printf("Ошибка запроса работ студентов: %v", err)
		return nil, status.Errorf(codes.Internal, "Ошибка сервера: %v", err)
	}
	defer rows.Close()

	resp := &pb.GetStudentWorksByTaskResponse{Works: []*pb.GetStudentWorksByTaskResponse_StudentWork{}}
	for rows.Next() {
		var work pb.GetStudentWorksByTaskResponse_StudentWork
		var assistantID sql.NullInt32
		var assistantPatronymic sql.NullString
		if err := rows.Scan(&work.Id, &work.StudentName, &work.StudentSurname, &work.StudentPatronymic, &work.StudentEmail, &work.Status,
			&assistantID, &work.AssistantName, &work.AssistantSurname, &assistantPatronymic, &work.StudentId); err != nil {
			log.Printf("Ошибка сканирования строки: %v", err)
			return nil, status.Errorf(codes.Internal, "Ошибка обработки данных: %v", err)
		}
		if assistantID.Valid {
			work.AssistantId = int32(assistantID.Int32)
		}
		if assistantPatronymic.Valid {
			work.AssistantPatronymic = assistantPatronymic.String
		}
		resp.Works = append(resp.Works, &work)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк: %v", err)
		return nil, status.Errorf(codes.Internal, "Ошибка обработки данных: %v", err)
	}
	return resp, nil
}
func (s *server) GetAssistantsByDiscipline(ctx context.Context, req *pb.GetAssistantsByDisciplineRequest) (*pb.GetAssistantsByDisciplineResponse, error) {
	query := `
		SELECT DISTINCT u.id, u.name, u.surname, COALESCE(u.patronymic, '')
		FROM users u
		JOIN users_in_groups ug ON u.id = ug.user_id
		JOIN groups_in_disciplines gd ON ug.group_id = gd.group_id
		WHERE gd.discipline_id = $1 AND u.role = 'assistant'
	`
	rows, err := s.db.QueryContext(ctx, query, req.DisciplineId)
	if err != nil {
		log.Printf("Failed to query assistants: %v", err)
		return &pb.GetAssistantsByDisciplineResponse{
			Error: fmt.Sprintf("Failed to query assistants: %v", err),
		}, nil
	}
	defer rows.Close()

	var assistants []*pb.GetAssistantsByDisciplineResponse_Assistant
	for rows.Next() {
		var id int32
		var name, surname, patronymic string
		if err := rows.Scan(&id, &name, &surname, &patronymic); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		assistants = append(assistants, &pb.GetAssistantsByDisciplineResponse_Assistant{
			Id:         id,
			Name:       name,
			Surname:    surname,
			Patronymic: patronymic,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &pb.GetAssistantsByDisciplineResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &pb.GetAssistantsByDisciplineResponse{
		Assistants: assistants,
	}, nil
}

func (s *server) AssignAssistantsToWorks(ctx context.Context, req *pb.AssignAssistantsToWorksRequest) (*pb.AssignAssistantsToWorksResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return &pb.AssignAssistantsToWorksResponse{
			Error: fmt.Sprintf("Failed to begin transaction: %v", err),
		}, nil
	}
	defer tx.Rollback()

	query := `UPDATE student_works SET assistant_id = $1 WHERE id = $2`
	for _, assignment := range req.Assignments {
		_, err := tx.ExecContext(ctx, query, assignment.AssistantId, assignment.WorkId)
		if err != nil {
			log.Printf("Failed to update assistant_id for work %d: %v", assignment.WorkId, err)
			return &pb.AssignAssistantsToWorksResponse{
				Error: fmt.Sprintf("Failed to update assistant_id for work %d: %v", assignment.WorkId, err),
			}, nil
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return &pb.AssignAssistantsToWorksResponse{
			Error: fmt.Sprintf("Failed to commit transaction: %v", err),
		}, nil
	}

	return &pb.AssignAssistantsToWorksResponse{
		Success: true,
	}, nil
}

func (s *server) GetTasksForSeminarist(ctx context.Context, req *pb.GetTasksForSeminaristRequest) (*pb.GetTasksForSeminaristResponse, error) {
	query := `
		SELECT DISTINCT t.id, t.title, t.deadline
		FROM tasks t
		JOIN groups_in_disciplines gd ON t.discipline_id = gd.discipline_id
		JOIN users_in_groups ug ON gd.group_id = ug.group_id
		WHERE ug.user_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, req.SeminaristId)
	if err != nil {
		log.Printf("Failed to query tasks: %v", err)
		return &pb.GetTasksForSeminaristResponse{
			Error: fmt.Sprintf("Failed to query tasks: %v", err),
		}, nil
	}
	defer rows.Close()

	var tasks []*pb.GetTasksForSeminaristResponse_Task
	for rows.Next() {
		var id int32
		var title string
		var deadline time.Time
		if err := rows.Scan(&id, &title, &deadline); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		tasks = append(tasks, &pb.GetTasksForSeminaristResponse_Task{
			Id:       id,
			Title:    title,
			Deadline: deadline.Format(time.RFC3339),
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &pb.GetTasksForSeminaristResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &pb.GetTasksForSeminaristResponse{
		Tasks: tasks,
	}, nil
}

func (s *server) GetStudentWorksForSeminarist(ctx context.Context, req *pb.GetStudentWorksForSeminaristRequest) (*pb.GetStudentWorksForSeminaristResponse, error) {
	query := `
		SELECT sw.id, t.title, sw.created_at, CONCAT(u.name, ' ', u.surname) AS student_name, sw.task_id
		FROM student_works sw
		JOIN tasks t ON sw.task_id = t.id
		JOIN users u ON sw.student_id = u.id
		WHERE sw.seminarist_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, req.SeminaristId)
	if err != nil {
		log.Printf("Failed to query student works: %v", err)
		return &pb.GetStudentWorksForSeminaristResponse{
			Error: fmt.Sprintf("Failed to query student works: %v", err),
		}, nil
	}
	defer rows.Close()

	var works []*pb.GetStudentWorksForSeminaristResponse_StudentWork
	for rows.Next() {
		var id int32
		var title, studentName string
		var createdAt time.Time
		var task_id int32
		if err := rows.Scan(&id, &title, &createdAt, &studentName, &task_id); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		works = append(works, &pb.GetStudentWorksForSeminaristResponse_StudentWork{
			Id:          id,
			Title:       title,
			CreatedAt:   createdAt.Format(time.RFC3339),
			StudentName: studentName,
			TaskId:      task_id,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &pb.GetStudentWorksForSeminaristResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &pb.GetStudentWorksForSeminaristResponse{
		Works: works,
	}, nil
}

// возвращает слайс работ из массивов состоящих из id работы, name, deadline
func (s *server) GetTasksForLector(ctx context.Context, req *pb.GetTasksForLectorRequest) (*pb.GetTasksForLectorResponse, error) {
	query := `SELECT id, title, deadline FROM tasks WHERE lector_id = $1`
	rows, err := s.db.QueryContext(ctx, query, req.LectorId)
	if err != nil {
		return &pb.GetTasksForLectorResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var task pb.Task
		if err := rows.Scan(&task.Id, &task.Title, &task.Deadline); err != nil {
			return &pb.GetTasksForLectorResponse{Error: err.Error()}, nil
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return &pb.GetTasksForLectorResponse{Error: err.Error()}, nil
	}
	return &pb.GetTasksForLectorResponse{Tasks: tasks}, nil
}

func (s *server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, req.TaskId)
	if err != nil {
		return &pb.DeleteTaskResponse{Error: err.Error()}, nil
	}
	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *server) SetTaskTitle(ctx context.Context, req *pb.SetTaskTitleRequest) (*pb.SetTaskTitleResponse, error) {
	query := `UPDATE tasks SET title = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.Title, req.TaskId)
	if err != nil {
		return &pb.SetTaskTitleResponse{Error: err.Error()}, nil
	}
	return &pb.SetTaskTitleResponse{Success: true}, nil
}

func (s *server) SetTaskDescription(ctx context.Context, req *pb.SetTaskDescriptionRequest) (*pb.SetTaskDescriptionResponse, error) {
	query := `UPDATE tasks SET description = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.Description, req.TaskId)
	if err != nil {
		return &pb.SetTaskDescriptionResponse{Error: err.Error()}, nil
	}
	return &pb.SetTaskDescriptionResponse{Success: true}, nil
}

func (s *server) SetTaskDeadline(ctx context.Context, req *pb.SetTaskDeadlineRequest) (*pb.SetTaskDeadlineResponse, error) {
	query := `UPDATE tasks SET deadline = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.Deadline, req.TaskId)
	if err != nil {
		return &pb.SetTaskDeadlineResponse{Error: err.Error()}, nil
	}
	return &pb.SetTaskDeadlineResponse{Success: true}, nil
}

func (s *server) CreateWork(ctx context.Context, req *pb.CreateWorkRequest) (*pb.CreateWorkResponse, error) {
	query := `INSERT INTO tasks (lector_id, group_id, title, description, deadline, discipline_id, content_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var taskID int32
	err := s.db.QueryRowContext(ctx, query, req.LectorId, req.GroupId, req.Title, req.Description, req.Deadline, req.DisciplineId, req.ContentUrl).Scan(&taskID)
	if err != nil {
		return &pb.CreateWorkResponse{Error: err.Error()}, nil
	}
	return &pb.CreateWorkResponse{TaskId: taskID}, nil
}

func (s *server) LoadTaskName(ctx context.Context, req *pb.LoadTaskNameRequest) (*pb.LoadTaskNameResponse, error) {
	query := `SELECT title FROM tasks WHERE id = $1`
	var title string
	err := s.db.QueryRowContext(ctx, query, req.TaskId).Scan(&title)
	if err != nil {
		return &pb.LoadTaskNameResponse{Error: err.Error()}, nil
	}
	return &pb.LoadTaskNameResponse{Title: title}, nil
}

func (s *server) LoadTaskDescription(ctx context.Context, req *pb.LoadTaskDescriptionRequest) (*pb.LoadTaskDescriptionResponse, error) {
	query := `SELECT description FROM tasks WHERE id = $1`
	var description string
	err := s.db.QueryRowContext(ctx, query, req.TaskId).Scan(&description)
	if err != nil {
		return &pb.LoadTaskDescriptionResponse{Error: err.Error()}, nil
	}
	return &pb.LoadTaskDescriptionResponse{Description: description}, nil
}

func (s *server) LoadTaskDeadline(ctx context.Context, req *pb.LoadTaskDeadlineRequest) (*pb.LoadTaskDeadlineResponse, error) {
	query := `SELECT deadline FROM tasks WHERE id = $1`
	var deadline string
	err := s.db.QueryRowContext(ctx, query, req.TaskId).Scan(&deadline)
	if err != nil {
		return &pb.LoadTaskDeadlineResponse{Error: err.Error()}, nil
	}
	return &pb.LoadTaskDeadlineResponse{Deadline: deadline}, nil
}

// получение групп лектора
func (s *server) GetGroups(ctx context.Context, req *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	var groups []*pb.GetGroupsResponse_Group
	query := `
        SELECT sg.id, sg.name
        FROM student_groups sg
        JOIN users_in_groups uig ON sg.id = uig.group_id
        WHERE uig.user_id = $1
    `
	rows, err := s.db.Query(query, req.LectorId)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var group pb.GetGroupsResponse_Group
		if err := rows.Scan(&group.Id, &group.Name); err != nil {
			return nil, fmt.Errorf("failed to scan group: %v", err)
		}
		groups = append(groups, &group)
	}

	return &pb.GetGroupsResponse{Groups: groups}, nil
}

// получение дисциплин лектор
func (s *server) GetDisciplines(ctx context.Context, req *pb.GetDisciplinesRequest) (*pb.GetDisciplinesResponse, error) {
	var disciplines []*pb.GetDisciplinesResponse_Discipline
	query := `
        SELECT id, name
        FROM disciplines
        WHERE lector_id = $1
    `
	rows, err := s.db.Query(query, req.LectorId)
	if err != nil {
		return nil, fmt.Errorf("failed to query disciplines: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var discipline pb.GetDisciplinesResponse_Discipline
		if err := rows.Scan(&discipline.Id, &discipline.Name); err != nil {
			return nil, fmt.Errorf("failed to scan discipline: %v", err)
		}
		disciplines = append(disciplines, &discipline)
	}

	return &pb.GetDisciplinesResponse{Disciplines: disciplines}, nil
}

func (s *server) GetTaskDetails(ctx context.Context, req *pb.GetTaskDetailsRequest) (*pb.GetTaskDetailsResponse, error) {
	query := `
		SELECT t.title, t.description, t.deadline, g.name AS group_name, d.name AS discipline_name,
		       u.name AS lector_name, u.surname AS lector_surname, u.patronymic AS lector_patronymic,
		       t.discipline_id, t.group_id
		FROM tasks t
		JOIN student_groups g ON t.group_id = g.id
		JOIN disciplines d ON t.discipline_id = d.id
		JOIN users u ON t.lector_id = u.id
		WHERE t.id = $1
	`
	var resp pb.GetTaskDetailsResponse
	var patronymic sql.NullString
	err := s.db.QueryRowContext(ctx, query, req.TaskId).Scan(
		&resp.Title, &resp.Description, &resp.Deadline, &resp.GroupName, &resp.DisciplineName,
		&resp.LectorName, &resp.LectorSurname, &patronymic, &resp.DisciplineId, &resp.GroupId,
	)
	if err == sql.ErrNoRows {
		resp.Error = fmt.Sprintf("Задание с ID %d не найдено", req.TaskId)
		return &resp, nil
	}
	if err != nil {
		log.Printf("Ошибка получения деталей задания %d: %v", req.TaskId, err)
		resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
		return &resp, nil
	}
	resp.LectorPatronymic = patronymic.String
	return &resp, nil
}

func (s *server) UpdateTaskGroupAndDiscipline(ctx context.Context, req *pb.UpdateTaskGroupAndDisciplineRequest) (*pb.UpdateTaskGroupAndDisciplineResponse, error) {
	query := `UPDATE tasks SET group_id = $1, discipline_id = $2 WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query, req.GroupId, req.DisciplineId, req.TaskId)
	if err != nil {
		return &pb.UpdateTaskGroupAndDisciplineResponse{Error: err.Error()}, nil
	}
	return &pb.UpdateTaskGroupAndDisciplineResponse{Success: true}, nil
}

func (s *server) ListTasksForStudent(ctx context.Context, req *pb.ListTasksForStudentRequest) (*pb.ListTasksForStudentResponse, error) {
	// Проверка контекста
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.Canceled, "Request canceled: %v", ctx.Err())
	}

	query := `
        SELECT t.id, t.title, t.deadline, COALESCE(w.status, 'pending') AS status
        FROM tasks t
        JOIN student_groups sg ON t.group_id = sg.id
        JOIN users_in_groups ug ON sg.id = ug.group_id
        LEFT JOIN student_works w ON t.id = w.task_id AND w.student_id = $1
        WHERE ug.user_id = $1`
	rows, err := s.db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		log.Printf("Ошибка запроса работ для student_id %d: %v", req.StudentId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var tasks []*pb.Tasks
	for rows.Next() {
		var task pb.Tasks
		var deadline time.Time
		if err := rows.Scan(&task.Id, &task.Title, &deadline, &task.Status); err != nil {
			log.Printf("Ошибка сканирования строки для student_id %d: %v", req.StudentId, err)
			return nil, status.Errorf(codes.Internal, "Ошибка обработки данных: %v", err)
		}
		// Преобразование deadline в строку (RFC3339)
		task.Deadline = deadline.Format(time.RFC3339)
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для student_id %d: %v", req.StudentId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка итерации данных: %v", err)
	}

	if len(tasks) == 0 {
		log.Printf("Нет работ для student_id %d", req.StudentId)
	}

	return &pb.ListTasksForStudentResponse{Tasks: tasks}, nil
}

func (s *server) ListWorksForStudent(ctx context.Context, req *pb.ListWorksForStudentRequest) (*pb.ListWorksForStudentResponse, error) {
	query := `
    SELECT w.id, t.title, t.deadline, w.status
    FROM student_works w
    JOIN tasks t ON w.task_id = t.id
    WHERE w.student_id = $1`
	rows, err := s.db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		log.Printf("Ошибка запроса работ для student_id %d: %v", req.StudentId, err)
		return &pb.ListWorksForStudentResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var works []*pb.Work
	for rows.Next() {
		var work pb.Work
		if err := rows.Scan(&work.Id, &work.Title, &work.Deadline, &work.Status); err != nil {
			log.Printf("Ошибка сканирования строки для student_id %d: %v", req.StudentId, err)
			return &pb.ListWorksForStudentResponse{Error: err.Error()}, nil
		}
		works = append(works, &work)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для student_id %d: %v", req.StudentId, err)
		return &pb.ListWorksForStudentResponse{Error: err.Error()}, nil
	}
	return &pb.ListWorksForStudentResponse{Works: works}, nil
}

func main() {

	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	// конвертируем порт из строки в число, чтобы работал sql.open
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT value: %v", err)
	}

	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	log.Printf("Trying to connect to: %s", connStr) // Для отладки
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Настройка сервера gRPC (остальной код остается без изменений)
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWorkServiceServer(s, &server{db: db})

	log.Println("WorkService starting on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
