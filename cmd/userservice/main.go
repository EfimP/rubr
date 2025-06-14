package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "rubr/proto/user"
	"strconv"
	"time"
)

type server struct {
	pb.UnimplementedUserServiceServer
	db *sql.DB
}

func (s *server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	if req.Name == "" || req.Surname == "" ||
		req.Patronymic == "" || req.Email == "" || req.Password == "" {
		return &pb.RegisterUserResponse{Error: "All fields must be filled"}, nil
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.RegisterUserResponse{Error: "Failed to hash password"}, nil
	}

	// Вставка в базу данных
	query := `INSERT INTO users (name, surname, patronymic, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int
	err = s.db.QueryRow(query, req.Name, req.Surname, req.Patronymic, req.Email, hashedPassword).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // Уникальность email
			return &pb.RegisterUserResponse{Error: "User already exists"}, nil
		}
		return &pb.RegisterUserResponse{Error: err.Error()}, nil
	}

	return &pb.RegisterUserResponse{UserId: strconv.Itoa(id)}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return &pb.LoginResponse{Error: "Email and password are required"}, nil
	}

	var id int
	var hashedPassword []byte
	var role string
	query := `SELECT id, password, role FROM users WHERE email = $1`
	err := s.db.QueryRow(query, req.Email).Scan(&id, &hashedPassword, &role)
	if err == sql.ErrNoRows {
		return &pb.LoginResponse{Error: "User not found"}, nil
	}
	if err != nil {
		return &pb.LoginResponse{Error: err.Error()}, nil
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(req.Password)); err != nil {
		return &pb.LoginResponse{Error: "Invalid password"}, nil
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен действителен 24 часа
	})
	tokenString, err := token.SignedString([]byte("rubroz")) // safety key
	if err != nil {
		return &pb.LoginResponse{Error: "Failed to generate token"}, nil
	}

	return &pb.LoginResponse{UserId: strconv.Itoa(id), Token: tokenString, Role: role}, nil
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
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{db: db})

	log.Println("UserService starting on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
