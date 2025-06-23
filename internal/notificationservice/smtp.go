package notificationservice

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	pb "rubr/proto/notification"
	"time"
)

const (
	smtpHost = "smtp.gmail.com"
	smtpPort = "587"
	smtpUser = "rubrnotifications@gmail.com"
	smtpPass = "RubrAVE0744"
)

// NotificationRequest представляет общий запрос для уведомлений
type NotificationRequest struct {
	UserID    int32
	Email     string
	Message   string
	CreatedAt time.Time
}

// SendTaskNotification отправляет уведомление студентам о новом задании
func (s *Server) SendTaskNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	if req.Email == "" || req.Message == "" {
		return &pb.NotificationResponse{Error: "email and message are required"}, nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Новое задание!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send task notification email to %s: %v", req.Email, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to send email: %v", err)}, nil
	}

	// Логируем отправку в таблицу notifications
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserId, "Новое задание: "+req.Message, req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserId, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to log notification: %v", err)}, nil
	}
	return &pb.NotificationResponse{}, nil
}

// SendRegistrationNotification отправляет письмо с информацией о новом аккаунте
func (s *Server) SendRegistrationNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	if req.Email == "" || req.Message == "" {
		return &pb.NotificationResponse{Error: "email and message are required"}, nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Регистрация аккаунта!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send registration email to %s: %v", req.Email, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to send email: %v", err)}, nil
	}

	// Логируем отправку
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserId, "Регистрация аккаунта", req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserId, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to log notification: %v", err)}, nil
	}
	return &pb.NotificationResponse{}, nil
}

// SendPasswordResetNotification отправляет письмо с временным паролем для сброса
func (s *Server) SendPasswordResetNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	if req.Email == "" || req.Message == "" {
		return &pb.NotificationResponse{Error: "email and message are required"}, nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Сброс пароля!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", req.Email, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to send email: %v", err)}, nil
	}

	// Логируем отправку
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserId, "Сброс пароля", req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserId, err)
		return &pb.NotificationResponse{Error: fmt.Sprintf("failed to log notification: %v", err)}, nil
	}
	return &pb.NotificationResponse{}, nil
}
