package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "rubr/proto/work"
	"strconv"
)

type server struct {
	pb.UnimplementedWorkServiceServer
	db *sql.DB
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

	log.Println("UserService starting on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
