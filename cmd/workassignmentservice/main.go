package main

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	pb "rubr/proto/workassignment"
)

type server struct {
	db *sql.DB
	pb.UnimplementedWorkAssignmentServiceServer
}

func (s *server) GetWorksForAssistant(ctx context.Context, req *pb.GetWorksForAssistantRequest) (*pb.GetWorksForAssistantResponse, error) {
	assistantID := req.AssistantId

	query := `
        SELECT 
            sw.id AS work_id, 
            t.id AS task_id, 
            t.title AS task_title, 
            u.id AS student_id, 
            u.email AS student_email, 
            u.name AS student_name, 
            u.surname AS student_surname, 
            u.patronymic AS student_patronymic
        FROM student_works sw
        JOIN tasks t ON sw.task_id = t.id
        JOIN users u ON sw.student_id = u.id
        WHERE sw.assistant_id = $1
    `
	rows, err := s.db.QueryContext(ctx, query, assistantID)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}
	defer rows.Close()

	var works []*pb.WorkAssignment
	for rows.Next() {
		var work pb.WorkAssignment
		err := rows.Scan(
			&work.WorkId,
			&work.TaskId,
			&work.TaskTitle,
			&work.StudentId,
			&work.StudentEmail,
			&work.StudentName,
			&work.StudentSurname,
			&work.StudentPatronymic,
		)
		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			return &pb.GetWorksForAssistantResponse{Error: "Ошибка обработки данных"}, nil
		}
		works = append(works, &work)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка после обработки строк: %v", err)
		return &pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}

	return &pb.GetWorksForAssistantResponse{Works: works}, nil
}

func (s *server) GetWorkDetails(ctx context.Context, req *pb.GetWorkDetailsRequest) (*pb.GetWorkDetailsResponse, error) {
	workID := req.WorkId

	query := `
        SELECT 
            sw.id AS work_id,
            t.title AS task_title,
            t.description AS task_description,
            t.deadline AS task_deadline,
            sw.created_at AS created_at,
            sw.status AS status,
            sw.content_url AS content_url
        FROM student_works sw
        JOIN tasks t ON sw.task_id = t.id
        WHERE sw.id = $1
    `
	var work pb.GetWorkDetailsResponse
	err := s.db.QueryRowContext(ctx, query, workID).Scan(
		&work.WorkId,
		&work.TaskTitle,
		&work.TaskDescription,
		&work.TaskDeadline,
		&work.CreatedAt,
		&work.Status,
		&work.ContentUrl,
	)
	if err == sql.ErrNoRows {
		return &pb.GetWorkDetailsResponse{Error: "Работа не найдена"}, nil
	}
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &pb.GetWorkDetailsResponse{Error: "Ошибка сервера"}, nil
	}

	return &work, nil
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
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWorkAssignmentServiceServer(s, &server{db: db})

	log.Println("WorkAssignmentService starting on :50054")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
