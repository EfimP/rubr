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
	pb "rubr/proto/grade"
	"strconv"
)

type server struct {
	pb.UnimplementedGradingServiceServer
	db *sql.DB
}

func (s *server) SetBlockingCriteriaMark(ctx context.Context, req *pb.SetBlockingCriteriaMarkRequest) (*pb.SetBlockingCriteriaMarkResponse, error) {
	query := `
        INSERT INTO student_criteria_marks (student_work_id, criteria_id, mark, comment)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (student_work_id, criteria_id) DO UPDATE 
        SET mark = EXCLUDED.mark, comment = EXCLUDED.comment`
	_, err := s.db.ExecContext(ctx, query, req.WorkId, req.CriterionId, req.Mark, req.Comment)
	if err != nil {
		return &pb.SetBlockingCriteriaMarkResponse{Error: err.Error()}, nil
	}
	return &pb.SetBlockingCriteriaMarkResponse{}, nil
}

func (s *server) SetMainCriteriaMark(ctx context.Context, req *pb.SetMainCriteriaMarkRequest) (*pb.SetMainCriteriaMarkResponse, error) {
	query := `
        INSERT INTO student_criteria_marks (student_work_id, criteria_id, mark, comment)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (student_work_id, criteria_id) DO UPDATE 
        SET mark = EXCLUDED.mark, comment = EXCLUDED.comment`
	_, err := s.db.ExecContext(ctx, query, req.WorkId, req.CriterionId, req.Mark, req.Comment)
	if err != nil {
		return &pb.SetMainCriteriaMarkResponse{Error: err.Error()}, nil
	}
	return &pb.SetMainCriteriaMarkResponse{}, nil
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

	// Настройка сервера gRPC
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGradingServiceServer(s, &server{db: db})

	log.Println("GradingService starting on: 50057")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
