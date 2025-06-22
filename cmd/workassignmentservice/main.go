package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"rubr/internal/workassignmentservice"
	Pb "rubr/proto/workassignment"
	"strconv"
)

func main() {

	workassignmentservice.InitS3Client()

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
	Pb.RegisterWorkAssignmentServiceServer(s, &workassignmentservice.Server{Db: db})

	log.Println("WorkAssignmentService starting on :50054")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
