package userservice

import (
	"context"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	pb "rubr/proto/user"
	"strconv"
)

func (s *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
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
	err = s.Db.QueryRow(query, req.Name, req.Surname, req.Patronymic, req.Email, hashedPassword).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // Уникальность email
			return &pb.RegisterUserResponse{Error: "User already exists"}, nil
		}
		return &pb.RegisterUserResponse{Error: err.Error()}, nil
	}

	return &pb.RegisterUserResponse{UserId: strconv.Itoa(id)}, nil
}
