package workservice

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	Pb "rubr/proto/work"
	"time"
)

func (s *Server) GetStudentsByGroupAndDiscipline(ctx context.Context, req *Pb.GetStudentsByGroupAndDisciplineRequest) (*Pb.GetStudentsByGroupAndDisciplineResponse, error) {
	resp := &Pb.GetStudentsByGroupAndDisciplineResponse{
		Students: make([]*Pb.GetStudentsByGroupAndDisciplineResponse_Student, 0),
	}

	// Проверяем связь group_id и discipline_id в groups_in_disciplines
	var exists bool
	err := s.Db.QueryRowContext(ctx, `
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
	rows, err := s.Db.QueryContext(ctx, query, req.GroupId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		resp.Error = fmt.Sprintf("Ошибка сервера: %v", err)
		return resp, nil
	}
	defer rows.Close()

	for rows.Next() {
		var student Pb.GetStudentsByGroupAndDisciplineResponse_Student
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

func (s *Server) GetStudentWorksByTask(ctx context.Context, req *Pb.GetStudentWorksByTaskRequest) (*Pb.GetStudentWorksByTaskResponse, error) {
	query := `
		SELECT sw.id, u.name, u.surname, u.patronymic, u.email, sw.status, sw.assistant_id,
		       COALESCE(a.name, '') AS assistant_name, COALESCE(a.surname, '') AS assistant_surname, COALESCE(a.patronymic, '') AS assistant_patronymic,
		       sw.student_id
		FROM student_works sw
		JOIN users u ON sw.student_id = u.id
		LEFT JOIN users a ON sw.assistant_id = a.id
		WHERE sw.task_id = $1
	`
	rows, err := s.Db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		log.Printf("Ошибка запроса работ студентов: %v", err)
		return nil, status.Errorf(codes.Internal, "Ошибка сервера: %v", err)
	}
	defer rows.Close()

	resp := &Pb.GetStudentWorksByTaskResponse{Works: []*Pb.GetStudentWorksByTaskResponse_StudentWork{}}
	for rows.Next() {
		var work Pb.GetStudentWorksByTaskResponse_StudentWork
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

func (s *Server) GetAssistantsByDiscipline(ctx context.Context, req *Pb.GetAssistantsByDisciplineRequest) (*Pb.GetAssistantsByDisciplineResponse, error) {
	query := `
		SELECT DISTINCT u.id, u.name, u.surname, COALESCE(u.patronymic, '')
		FROM users u
		JOIN users_in_groups ug ON u.id = ug.user_id
		JOIN groups_in_disciplines gd ON ug.group_id = gd.group_id
		WHERE gd.discipline_id = $1 AND u.role = 'assistant'
	`
	rows, err := s.Db.QueryContext(ctx, query, req.DisciplineId)
	if err != nil {
		log.Printf("Failed to query assistants: %v", err)
		return &Pb.GetAssistantsByDisciplineResponse{
			Error: fmt.Sprintf("Failed to query assistants: %v", err),
		}, nil
	}
	defer rows.Close()

	var assistants []*Pb.GetAssistantsByDisciplineResponse_Assistant
	for rows.Next() {
		var id int32
		var name, surname, patronymic string
		if err := rows.Scan(&id, &name, &surname, &patronymic); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		assistants = append(assistants, &Pb.GetAssistantsByDisciplineResponse_Assistant{
			Id:         id,
			Name:       name,
			Surname:    surname,
			Patronymic: patronymic,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &Pb.GetAssistantsByDisciplineResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &Pb.GetAssistantsByDisciplineResponse{
		Assistants: assistants,
	}, nil
}

func (s *Server) GetTasksForSeminarist(ctx context.Context, req *Pb.GetTasksForSeminaristRequest) (*Pb.GetTasksForSeminaristResponse, error) {
	query := `
		SELECT DISTINCT t.id, t.title, t.deadline
		FROM tasks t
		JOIN groups_in_disciplines gd ON t.discipline_id = gd.discipline_id
		JOIN users_in_groups ug ON gd.group_id = ug.group_id
		WHERE ug.user_id = $1
	`
	rows, err := s.Db.QueryContext(ctx, query, req.SeminaristId)
	if err != nil {
		log.Printf("Failed to query tasks: %v", err)
		return &Pb.GetTasksForSeminaristResponse{
			Error: fmt.Sprintf("Failed to query tasks: %v", err),
		}, nil
	}
	defer rows.Close()

	var tasks []*Pb.GetTasksForSeminaristResponse_Task
	for rows.Next() {
		var id int32
		var title string
		var deadline time.Time
		if err := rows.Scan(&id, &title, &deadline); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		tasks = append(tasks, &Pb.GetTasksForSeminaristResponse_Task{
			Id:       id,
			Title:    title,
			Deadline: deadline.Format(time.RFC3339),
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &Pb.GetTasksForSeminaristResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &Pb.GetTasksForSeminaristResponse{
		Tasks: tasks,
	}, nil
}

func (s *Server) GetStudentWorksForSeminarist(ctx context.Context, req *Pb.GetStudentWorksForSeminaristRequest) (*Pb.GetStudentWorksForSeminaristResponse, error) {
	query := `
		SELECT sw.id, t.title, sw.created_at, CONCAT(u.name, ' ', u.surname) AS student_name, sw.task_id
		FROM student_works sw
		JOIN tasks t ON sw.task_id = t.id
		JOIN users u ON sw.student_id = u.id
		WHERE sw.seminarist_id = $1
	`
	rows, err := s.Db.QueryContext(ctx, query, req.SeminaristId)
	if err != nil {
		log.Printf("Failed to query student works: %v", err)
		return &Pb.GetStudentWorksForSeminaristResponse{
			Error: fmt.Sprintf("Failed to query student works: %v", err),
		}, nil
	}
	defer rows.Close()

	var works []*Pb.GetStudentWorksForSeminaristResponse_StudentWork
	for rows.Next() {
		var id int32
		var title, studentName string
		var createdAt time.Time
		var task_id int32
		if err := rows.Scan(&id, &title, &createdAt, &studentName, &task_id); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		works = append(works, &Pb.GetStudentWorksForSeminaristResponse_StudentWork{
			Id:          id,
			Title:       title,
			CreatedAt:   createdAt.Format(time.RFC3339),
			StudentName: studentName,
			TaskId:      task_id,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return &Pb.GetStudentWorksForSeminaristResponse{
			Error: fmt.Sprintf("Row iteration error: %v", err),
		}, nil
	}

	return &Pb.GetStudentWorksForSeminaristResponse{
		Works: works,
	}, nil
}

// возвращает слайс работ из массивов состоящих из id работы, name, deadline
func (s *Server) GetTasksForLector(ctx context.Context, req *Pb.GetTasksForLectorRequest) (*Pb.GetTasksForLectorResponse, error) {
	query := `SELECT id, title, deadline FROM tasks WHERE lector_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, req.LectorId)
	if err != nil {
		return &Pb.GetTasksForLectorResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var tasks []*Pb.Task
	for rows.Next() {
		var task Pb.Task
		if err := rows.Scan(&task.Id, &task.Title, &task.Deadline); err != nil {
			return &Pb.GetTasksForLectorResponse{Error: err.Error()}, nil
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return &Pb.GetTasksForLectorResponse{Error: err.Error()}, nil
	}
	return &Pb.GetTasksForLectorResponse{Tasks: tasks}, nil
}

// получение групп лектора
func (s *Server) GetGroups(ctx context.Context, req *Pb.GetGroupsRequest) (*Pb.GetGroupsResponse, error) {
	var groups []*Pb.GetGroupsResponse_Group
	query := `
        SELECT sg.id, sg.name
        FROM student_groups sg
        JOIN users_in_groups uig ON sg.id = uig.group_id
        WHERE uig.user_id = $1
    `
	rows, err := s.Db.Query(query, req.LectorId)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var group Pb.GetGroupsResponse_Group
		if err := rows.Scan(&group.Id, &group.Name); err != nil {
			return nil, fmt.Errorf("failed to scan group: %v", err)
		}
		groups = append(groups, &group)
	}

	return &Pb.GetGroupsResponse{Groups: groups}, nil
}

// получение дисциплин лектор
func (s *Server) GetDisciplines(ctx context.Context, req *Pb.GetDisciplinesRequest) (*Pb.GetDisciplinesResponse, error) {
	var disciplines []*Pb.GetDisciplinesResponse_Discipline
	query := `
        SELECT id, name
        FROM disciplines
        WHERE lector_id = $1
    `
	rows, err := s.Db.Query(query, req.LectorId)
	if err != nil {
		return nil, fmt.Errorf("failed to query disciplines: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var discipline Pb.GetDisciplinesResponse_Discipline
		if err := rows.Scan(&discipline.Id, &discipline.Name); err != nil {
			return nil, fmt.Errorf("failed to scan discipline: %v", err)
		}
		disciplines = append(disciplines, &discipline)
	}

	return &Pb.GetDisciplinesResponse{Disciplines: disciplines}, nil
}

func (s *Server) GetStudentDisciplines(ctx context.Context, req *Pb.GetStudentDisciplinesRequest) (*Pb.GetStudentDisciplinesResponse, error) {
	if ctx.Err() != nil {
		return &Pb.GetStudentDisciplinesResponse{Error: "Request canceled"}, nil
	}

	query := `
		SELECT DISTINCT d.id, d.name
		FROM disciplines d
		JOIN groups_in_disciplines gd ON d.id = gd.discipline_id
		JOIN users_in_groups ug ON gd.group_id = ug.group_id
		WHERE ug.user_id = $1
	`
	rows, err := s.Db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		log.Printf("Ошибка получения дисциплин для student_id %d: %v", req.StudentId, err)
		return &Pb.GetStudentDisciplinesResponse{Error: "Ошибка сервера"}, nil
	}
	defer rows.Close()

	var disciplines []*Pb.GetStudentDisciplinesResponse_Discipline
	for rows.Next() {
		var discipline Pb.GetStudentDisciplinesResponse_Discipline
		if err := rows.Scan(&discipline.Id, &discipline.Name); err != nil {
			log.Printf("Ошибка сканирования строки для student_id %d: %v", req.StudentId, err)
			return &Pb.GetStudentDisciplinesResponse{Error: "Ошибка обработки данных"}, nil
		}
		disciplines = append(disciplines, &discipline)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для student_id %d: %v", req.StudentId, err)
		return &Pb.GetStudentDisciplinesResponse{Error: "Ошибка обработки данных"}, nil
	}

	return &Pb.GetStudentDisciplinesResponse{Disciplines: disciplines}, nil
}

func (s *Server) GetStudentWorksByDiscipline(ctx context.Context, req *Pb.GetStudentWorksByDisciplineRequest) (*Pb.GetStudentWorksByDisciplineResponse, error) {
	if ctx.Err() != nil {
		return &Pb.GetStudentWorksByDisciplineResponse{Error: "Request canceled"}, nil
	}

	// Отладочный лог для входных данных
	log.Printf("Запрос работ для student_id=%d, discipline_id=%d в %s", req.StudentId, req.DisciplineId, time.Now().Format("15:04:05 02-01-2006"))

	// Запрос для получения уникальных работ по дисциплине
	query := `
		SELECT DISTINCT sw.id, t.title, t.deadline, sw.status
		FROM student_works sw
		JOIN tasks t ON sw.task_id = t.id
		JOIN groups_in_disciplines gd ON t.group_id = gd.group_id AND t.discipline_id = gd.discipline_id
		JOIN users_in_groups ug ON gd.group_id = ug.group_id AND ug.user_id = sw.student_id
		WHERE sw.student_id = $1 AND t.discipline_id = $2
	`
	rows, err := s.Db.QueryContext(ctx, query, req.StudentId, req.DisciplineId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса для student_id=%d, discipline_id=%d: %v", req.StudentId, req.DisciplineId, err)
		return &Pb.GetStudentWorksByDisciplineResponse{Error: "Ошибка сервера"}, nil
	}
	defer rows.Close()

	var works []*Pb.Work
	workIDs := make(map[int32]bool) // Для дополнительной проверки уникальности
	for rows.Next() {
		var work Pb.Work
		var deadline time.Time
		if err := rows.Scan(&work.Id, &work.Title, &deadline, &work.Status); err != nil {
			log.Printf("Ошибка сканирования строки для student_id=%d: %v", req.StudentId, err)
			return &Pb.GetStudentWorksByDisciplineResponse{Error: "Ошибка обработки данных"}, nil
		}
		work.Deadline = deadline.Format(time.RFC3339)

		// Проверка на дубликат
		if !workIDs[work.Id] {
			works = append(works, &work)
			workIDs[work.Id] = true
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для student_id=%d: %v", req.StudentId, err)
		return &Pb.GetStudentWorksByDisciplineResponse{Error: "Ошибка обработки данных"}, nil
	}

	log.Printf("Найдено %d уникальных работ для student_id=%d, discipline_id=%d: %v", len(works), req.StudentId, req.DisciplineId, works)
	return &Pb.GetStudentWorksByDisciplineResponse{Works: works}, nil
}

func (s *Server) GetTaskDetails(ctx context.Context, req *Pb.GetTaskDetailsRequest) (*Pb.GetTaskDetailsResponse, error) {
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
	var resp Pb.GetTaskDetailsResponse
	var patronymic sql.NullString
	err := s.Db.QueryRowContext(ctx, query, req.TaskId).Scan(
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

func (s *Server) ListTasksForStudent(ctx context.Context, req *Pb.ListTasksForStudentRequest) (*Pb.ListTasksForStudentResponse, error) {
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
	rows, err := s.Db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		log.Printf("Ошибка запроса работ для student_id %d: %v", req.StudentId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var tasks []*Pb.Tasks
	for rows.Next() {
		var task Pb.Tasks
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

	return &Pb.ListTasksForStudentResponse{Tasks: tasks}, nil
}

func (s *Server) ListWorksForStudent(ctx context.Context, req *Pb.ListWorksForStudentRequest) (*Pb.ListWorksForStudentResponse, error) {
	query := `
    SELECT w.id, t.title, t.deadline, w.status
    FROM student_works w
    JOIN tasks t ON w.task_id = t.id
    WHERE w.student_id = $1`
	rows, err := s.Db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		log.Printf("Ошибка запроса работ для student_id %d: %v", req.StudentId, err)
		return &Pb.ListWorksForStudentResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var works []*Pb.Work
	for rows.Next() {
		var work Pb.Work
		if err := rows.Scan(&work.Id, &work.Title, &work.Deadline, &work.Status); err != nil {
			log.Printf("Ошибка сканирования строки для student_id %d: %v", req.StudentId, err)
			return &Pb.ListWorksForStudentResponse{Error: err.Error()}, nil
		}
		works = append(works, &work)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для student_id %d: %v", req.StudentId, err)
		return &Pb.ListWorksForStudentResponse{Error: err.Error()}, nil
	}
	return &Pb.ListWorksForStudentResponse{Works: works}, nil
}
