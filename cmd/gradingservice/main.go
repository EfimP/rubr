package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	pb "rubr/proto/grade"
	"strconv"
	"strings"
)

type server struct {
	pb.UnimplementedGradingServiceServer
	db *sql.DB
}

func (s *server) UpdateWorkStatus(ctx context.Context, req *pb.UpdateWorkStatusRequest) (*pb.UpdateWorkStatusResponse, error) {
	log.Printf("Получен запрос UpdateWorkStatus для work_id: %d, status: %s", req.WorkId, req.Status)

	// Валидация входных данных
	if req.WorkId <= 0 {
		log.Printf("Неверный work_id: %d", req.WorkId)
		return nil, status.Errorf(codes.InvalidArgument, "work_id должен быть положительным")
	}
	if req.Status == "" {
		log.Printf("Пустой статус для work_id: %d", req.WorkId)
		return nil, status.Errorf(codes.InvalidArgument, "status не должен быть пустым")
	}

	// SQL-запрос для обновления статуса
	query := `
        UPDATE student_works
        SET status = $1
        WHERE id = $2
        RETURNING id`
	var updatedID int64
	err := s.db.QueryRowContext(ctx, query, req.Status, req.WorkId).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Работа с id %d не найдена", req.WorkId)
			return &pb.UpdateWorkStatusResponse{Error: fmt.Sprintf("работа с id %d не найдена", req.WorkId)}, nil
		}
		log.Printf("Ошибка обновления статуса для work_id %d: %v", req.WorkId, err)
		return nil, status.Errorf(codes.Internal, "ошибка базы данных: %v", err)
	}

	log.Printf("Статус работы %d успешно обновлен на %s", req.WorkId, req.Status)
	return &pb.UpdateWorkStatusResponse{}, nil
}

func (s *server) GetCriteriaMarks(ctx context.Context, req *pb.GetCriteriaMarksRequest) (*pb.GetCriteriaMarksResponse, error) {
	log.Printf("Получен запрос GetCriteriaMarks для work_id: %d", req.WorkId)

	// Проверка входных данных
	if req.WorkId <= 0 {
		log.Printf("Неверный work_id: %d", req.WorkId)
		return &pb.GetCriteriaMarksResponse{Error: "work_id должен быть положительным"}, nil
	}

	// SQL-запрос для получения оценок
	query := `
        SELECT criteria_id, mark, COALESCE(comment, '')
        FROM student_criteria_marks
        WHERE student_work_id = $1`
	rows, err := s.db.QueryContext(ctx, query, req.WorkId)
	if err != nil {
		log.Printf("Ошибка выполнения запроса для work_id %d: %v", req.WorkId, err)
		return &pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка базы данных: %v", err)}, nil
	}
	defer rows.Close()

	// Сбор результатов
	var marks []*pb.CriterionMark
	for rows.Next() {
		var criterionID int64
		var mark float32
		var comment string
		if err := rows.Scan(&criterionID, &mark, &comment); err != nil {
			log.Printf("Ошибка сканирования строки для work_id %d: %v", req.WorkId, err)
			return &pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка обработки данных: %v", err)}, nil
		}
		marks = append(marks, &pb.CriterionMark{
			CriterionId: int32(criterionID),
			Mark:        mark,
			Comment:     comment,
		})
	}

	// Проверка ошибок после итерации
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для work_id %d: %v", req.WorkId, err)
		return &pb.GetCriteriaMarksResponse{Error: fmt.Sprintf("ошибка обработки данных: %v", err)}, nil
	}

	log.Printf("Найдено %d оценок для work_id %d", len(marks), req.WorkId)
	return &pb.GetCriteriaMarksResponse{
		Marks: marks,
	}, nil
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

func (s *server) ListSubjects(ctx context.Context, req *pb.ListSubjectsRequest) (*pb.ListSubjectsResponse, error) {
	query := `SELECT name, grades, average FROM student_subjects WHERE student_id = $1`
	rows, err := s.db.QueryContext(ctx, query, req.StudentId)
	if err != nil {
		return &pb.ListSubjectsResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var subjects []*pb.Subject
	for rows.Next() {
		var subject pb.Subject
		var gradesStr string
		if err := rows.Scan(&subject.Name, &gradesStr, &subject.Average); err != nil {
			return &pb.ListSubjectsResponse{Error: err.Error()}, nil
		}
		// Парсинг grades (предполагается, что grades хранится как строка, например, "4.0,3.5,4.5")
		for _, g := range strings.Split(gradesStr, ",") {
			grade, _ := strconv.ParseFloat(g, 32)
			subject.Grades = append(subject.Grades, float32(grade))
		}
		subjects = append(subjects, &subject)
	}
	if err := rows.Err(); err != nil {
		return &pb.ListSubjectsResponse{Error: err.Error()}, nil
	}
	return &pb.ListSubjectsResponse{Subjects: subjects}, nil
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
