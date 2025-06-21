package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"rubr/internal/gradingservice"
	Pb "rubr/proto/grade"
	"strconv"
)

func main() {

	DbHost := os.Getenv("Db_HOST")
	DbPortStr := os.Getenv("Db_PORT")
	DbUser := os.Getenv("Db_USER")
	DbPassword := os.Getenv("Db_PASSWORD")
	DbName := os.Getenv("Db_NAME")
	// конвертируем порт из строки в число, чтобы работал sql.open
	DbPort, err := strconv.Atoi(DbPortStr)
	if err != nil {
		log.Fatalf("Invalid Db_PORT value: %v", err)
	}

	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s Dbname=%s sslmode=disable",
		DbHost, DbPort, DbUser, DbPassword, DbName)
	log.Printf("Trying to connect to: %s", connStr) // Для отладки
	Db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer Db.Close()

	// Настройка сервера gRPC
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	Pb.RegisterGradingServiceServer(s, &gradingservice.Server{Db: Db})

	log.Println("GradingService starting on: 50057")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
