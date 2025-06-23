package userservice

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"log"
	pb "rubr/proto/user"
)

// UpdatePassword обновляет пароль пользователя по его email
func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	if req.Email == "" || req.Password == "" {
		return &pb.UpdatePasswordResponse{Error: "email and password are required"}, nil
	}

	// Хешируем пароль с использованием bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.UpdatePasswordResponse{Error: "Failed to hash password"}, nil
	}

	if err != nil {
		log.Printf("Failed to hash password for email %s: %v", req.Email, err)
		return &pb.UpdatePasswordResponse{Error: "internal server error"}, nil
	}

	// Обновляем пароль в базе данных
	result, err := s.Db.ExecContext(ctx, `
		UPDATE users
		SET password = $1
		WHERE email = $2`,
		hashedPassword, req.Email)
	if err != nil {
		log.Printf("Failed to update password for email %s: %v", req.Email, err)
		return &pb.UpdatePasswordResponse{Error: "failed to update password"}, nil
	}

	// Проверяем, было ли обновление
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check rows affected for email %s: %v", req.Email, err)
		return &pb.UpdatePasswordResponse{Error: "internal server error"}, nil
	}
	if rowsAffected == 0 {
		return &pb.UpdatePasswordResponse{Error: "user not found"}, nil
	}

	return &pb.UpdatePasswordResponse{}, nil
}
