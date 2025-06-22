package notificationservice

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

const (
	smtpHost = "smtp.gmail.com"
	smtpPort = "587"
	smtpUser = "your-email@gmail.com"
	smtpPass = "your-app-password"
)

// NotificationRequest представляет общий запрос для уведомлений
type NotificationRequest struct {
	UserID    int32
	Email     string
	Message   string
	CreatedAt time.Time
}

// SendTaskNotification отправляет уведомление студентам о новом задании
func (s *Server) SendTaskNotification(ctx context.Context, req *NotificationRequest) error {
	if req.Email == "" || req.Message == "" {
		return fmt.Errorf("email and message are required")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Новое задание!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send task notification email to %s: %v", req.Email, err)
		return err
	}

	// Логируем отправку в таблицу notifications
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserID, "Новое задание: "+req.Message, req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserID, err)
	}
	return nil
}

// SendRegistrationNotification отправляет письмо с информацией о новом аккаунте
func (s *Server) SendRegistrationNotification(ctx context.Context, req *NotificationRequest) error {
	if req.Email == "" || req.Message == "" {
		return fmt.Errorf("email and message are required")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Регистрация аккаунта!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send registration email to %s: %v", req.Email, err)
		return err
	}

	// Логируем отправку
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserID, "Регистрация аккаунта", req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserID, err)
	}
	return nil
}

// SendPasswordResetNotification отправляет письмо с временным паролем для сброса
func (s *Server) SendPasswordResetNotification(ctx context.Context, req *NotificationRequest) error {
	// Проверяем, что данные переданы
	if req.Email == "" || req.Message == "" {
		return fmt.Errorf("email and message are required")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{req.Email}, []byte("Subject: Сброс пароля!\n\n"+req.Message))
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", req.Email, err)
		return err
	}

	// Логируем отправку
	_, err = s.Db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)`,
		req.UserID, "Сброс пароля", req.CreatedAt)
	if err != nil {
		log.Printf("Failed to log notification for user %d: %v", req.UserID, err)
	}
	return nil
}
