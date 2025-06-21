package userservice

import (
	"context"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	pb "rubr/proto/user"
	"strconv"
	"time"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return &pb.LoginResponse{Error: "Email and password are required"}, nil
	}

	var id int
	var hashedPassword []byte
	var role string
	query := `SELECT id, password, role FROM users WHERE email = $1`
	err := s.Db.QueryRow(query, req.Email).Scan(&id, &hashedPassword, &role)
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
