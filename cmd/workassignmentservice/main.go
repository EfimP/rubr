package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	pb "rubr/proto/workassignment"
	"strconv"
	"time"
)

type server struct {
	pb.UnimplementedWorkAssignmentServiceServer
	db *sql.DB
}

var S3Client *s3.Client

func initS3Client() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Ошибка инициализации S3 клиента: %v", err)
	}
	S3Client = s3.NewFromConfig(cfg)
}

func (s *server) GetWorksForAssistant(ctx context.Context, req *pb.GetWorksForAssistantRequest) (*pb.GetWorksForAssistantResponse, error) {
	assistantID := req.AssistantId

	query := `
        SELECT 
            sw.id AS work_id, 
            t.id AS task_id, 
            t.title AS task_title, 
            u.id AS student_id, 
            u.email AS student_email, 
            u.name AS student_name, 
            u.surname AS student_surname, 
            u.patronymic AS student_patronymic
        FROM student_works sw
        JOIN tasks t ON sw.task_id = t.id
        JOIN users u ON sw.student_id = u.id
        WHERE sw.assistant_id = $1
    `
	rows, err := s.db.QueryContext(ctx, query, assistantID)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}
	defer rows.Close()

	var works []*pb.WorkAssignment
	for rows.Next() {
		var work pb.WorkAssignment
		err := rows.Scan(
			&work.WorkId,
			&work.TaskId,
			&work.TaskTitle,
			&work.StudentId,
			&work.StudentEmail,
			&work.StudentName,
			&work.StudentSurname,
			&work.StudentPatronymic,
		)
		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			return &pb.GetWorksForAssistantResponse{Error: "Ошибка обработки данных"}, nil
		}
		works = append(works, &work)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка после обработки строк: %v", err)
		return &pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}

	return &pb.GetWorksForAssistantResponse{Works: works}, nil
}

func (s *server) GetWorkDetails(ctx context.Context, req *pb.GetWorkDetailsRequest) (*pb.GetWorkDetailsResponse, error) {
	workID := req.WorkId

	query := `
        SELECT 
            sw.id AS work_id,
            t.title AS task_title,
            t.description AS task_description,
            t.deadline AS task_deadline,
            sw.created_at AS created_at,
            sw.status AS status,
            sw.content_url AS content_url
        FROM student_works sw
        JOIN tasks t ON sw.task_id = t.id
        WHERE sw.id = $1
    `
	var work pb.GetWorkDetailsResponse
	err := s.db.QueryRowContext(ctx, query, workID).Scan(
		&work.WorkId,
		&work.TaskTitle,
		&work.TaskDescription,
		&work.TaskDeadline,
		&work.CreatedAt,
		&work.Status,
		&work.ContentUrl,
	)
	if err == sql.ErrNoRows {
		return &pb.GetWorkDetailsResponse{Error: "Работа не найдена"}, nil
	}
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &pb.GetWorkDetailsResponse{Error: "Ошибка сервера"}, nil
	}

	return &work, nil
}

func (s *server) GetTaskDetails(ctx context.Context, req *pb.GetTaskDetailsRequest) (*pb.GetTaskDetailsResponse, error) {
	// Проверка контекста
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.Canceled, "Request canceled: %v", ctx.Err())
	}

	query := `
        SELECT 
            t.id AS task_id,
            t.title AS task_title,
            t.description AS task_description,
            t.deadline AS task_deadline,
            t.content_url AS task_content_url,
            cg.id AS criteria_group_id,
            cg.group_name AS criteria_group_name,
            cg.block_flag AS criteria_block_flag
        FROM tasks t
        LEFT JOIN criteria_groups cg ON t.id = cg.task_id
        WHERE t.id = $1`

	rows, err := s.db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных для task_id %d: %v", req.TaskId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка сервера: %v", err)
	}
	defer rows.Close()

	var response pb.GetTaskDetailsResponse
	var taskDeadline time.Time
	for rows.Next() {
		var taskID int32
		var criteriaGroupID int32
		var criteriaGroupName string
		var criteriaBlockFlag bool
		if err := rows.Scan(
			&taskID,
			&response.TaskTitle,
			&response.TaskDescription,
			&taskDeadline,
			&response.TaskContentUrl,
			&criteriaGroupID,
			&criteriaGroupName,
			&criteriaBlockFlag,
		); err != nil {
			log.Printf("Ошибка сканирования строки для task_id %d: %v", req.TaskId, err)
			return nil, status.Errorf(codes.Internal, "Ошибка обработки данных: %v", err)
		}

		// Устанавливаем TaskId, если это первая итерация
		if response.TaskId == 0 {
			response.TaskId = taskID
			response.TaskDeadline = taskDeadline.Format(time.RFC3339)
		}
		// Добавляем группу критериев в соответствующий список
		cg := &pb.CriteriaGroup{
			Id:         criteriaGroupID,
			Name:       criteriaGroupName,
			IsBlocking: criteriaBlockFlag,
		}
		if criteriaBlockFlag {
			response.BlockingCriteriaGroups = append(response.BlockingCriteriaGroups, cg)
		} else {
			response.MainCriteriaGroups = append(response.MainCriteriaGroups, cg)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации строк для task_id %d: %v", req.TaskId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка итерации данных: %v", err)
	}
	return &response, nil
}
func (s *server) UploadAssignmentFile(ctx context.Context, req *pb.UploadAssignmentFileRequest) (*pb.UploadAssignmentFileResponse, error) {
	query := `UPDATE student_works SET content_url = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.FilePath, req.WorkId)
	if err != nil {
		return &pb.UploadAssignmentFileResponse{Error: err.Error()}, nil
	}
	return &pb.UploadAssignmentFileResponse{Success: true}, nil
}

func (s *server) CheckExistingWork(ctx context.Context, req *pb.CheckExistingWorkRequest) (*pb.CheckExistingWorkResponse, error) {
	if ctx.Err() != nil {
		return &pb.CheckExistingWorkResponse{Error: "Request canceled"}, nil
	}

	var exists bool
	var workID int64
	err := s.db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM student_works WHERE student_id = $1 AND task_id = $2),
               id
        FROM student_works 
        WHERE student_id = $1 AND task_id = $2`,
		req.StudentId, req.TaskId).Scan(&exists, &workID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Ошибка проверки работы для student_id %d и task_id %d: %v", req.StudentId, req.TaskId, err)
		return &pb.CheckExistingWorkResponse{Error: "Ошибка сервера"}, nil
	}

	if !exists {
		return &pb.CheckExistingWorkResponse{Exists: false, StudentId: req.StudentId}, nil
	}

	return &pb.CheckExistingWorkResponse{
		Exists:    true,
		WorkId:    int32(workID),
		StudentId: req.StudentId,
	}, nil
}

func (s *server) SubmitWork(ctx context.Context, req *pb.SubmitWorkRequest) (*pb.SubmitWorkResponse, error) {
	query := `UPDATE student_works SET status = 'submitted', content_url = $1, created_at = CURRENT_TIMESTAMP WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, req.FilePath, req.WorkId)
	if err != nil {
		log.Printf("Ошибка обновления работы %d: %v", req.WorkId, err)
		return &pb.SubmitWorkResponse{Error: err.Error()}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Printf("Работа %d не найдена или не обновлена", req.WorkId)
		return &pb.SubmitWorkResponse{Error: "Работа не найдена"}, nil
	}
	log.Printf("Работа %d успешно сдана", req.WorkId)
	return &pb.SubmitWorkResponse{Success: true}, nil
}

func (s *server) CreateWork(ctx context.Context, req *pb.CreateWorkRequest) (*pb.CreateWorkResponse, error) {
	if ctx.Err() != nil {
		return &pb.CreateWorkResponse{Error: "Request canceled"}, nil
	}

	// Начало транзакции
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	defer tx.Rollback()

	var studentExists, taskExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.StudentId).Scan(&studentExists)
	if err != nil {
		log.Printf("Ошибка проверки студента %d: %v", req.StudentId, err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	if !studentExists {
		return &pb.CreateWorkResponse{Error: "Студент не найден"}, nil
	}

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", req.TaskId).Scan(&taskExists)
	if err != nil {
		log.Printf("Ошибка проверки задания %d: %v", req.TaskId, err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	if !taskExists {
		return &pb.CreateWorkResponse{Error: "Задание не найдено"}, nil
	}

	// Извлечение group_id из tasks
	var groupID int64
	err = tx.QueryRowContext(ctx, `
        SELECT group_id FROM tasks WHERE id = $1`, req.TaskId).Scan(&groupID)
	if err != nil {
		log.Printf("Ошибка получения group_id для task_id %d: %v", req.TaskId, err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}

	// Извлечение seminarist_id из groups_in_disciplines
	var seminaristID int64
	err = tx.QueryRowContext(ctx, `
        SELECT seminarist_id FROM groups_in_disciplines WHERE group_id = $1`, groupID).Scan(&seminaristID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Не найден seminarist_id для group_id %d", groupID)
			return &pb.CreateWorkResponse{Error: "Семинарист не назначен для группы"}, nil
		}
		log.Printf("Ошибка получения seminarist_id для group_id %d: %v", groupID, err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}

	// Проверка существования работы и обновление/создание
	var workID int64
	err = tx.QueryRowContext(ctx, `
        SELECT id FROM student_works WHERE student_id = $1 AND task_id = $2 FOR UPDATE`,
		req.StudentId, req.TaskId).Scan(&workID)
	if err == sql.ErrNoRows {
		// Если работы нет, создаем новую
		err = tx.QueryRowContext(ctx, `
            INSERT INTO student_works (student_id, task_id, status, seminarist_id)
            VALUES ($1, $2, 'submitted', $3)
            RETURNING id`,
			req.StudentId, req.TaskId, seminaristID).Scan(&workID)
		if err != nil {
			log.Printf("Ошибка создания работы для student_id %d и task_id %d: %v", req.StudentId, req.TaskId, err)
			return &pb.CreateWorkResponse{Error: "Ошибка создания работы"}, nil
		}
		log.Printf("Создано новая работа с ID %d для student_id %d, task_id %d, seminarist_id %d", workID, req.StudentId, req.TaskId, seminaristID)
	} else if err != nil {
		log.Printf("Ошибка проверки существующей работы для student_id %d и task_id %d: %v", req.StudentId, req.TaskId, err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	} else {
		// Если работа существует, обновляем данные
		_, err = tx.ExecContext(ctx, `
            UPDATE student_works 
            SET status = 'submitted', seminarist_id = $2
            WHERE id = $1`,
			workID, seminaristID)
		if err != nil {
			log.Printf("Ошибка обновления работы %d для student_id %d и task_id %d: %v", workID, req.StudentId, req.TaskId, err)
			return &pb.CreateWorkResponse{Error: fmt.Sprintf("Ошибка обновления работы: %v", err)}, nil
		}
		log.Printf("Обновлена работа с ID %d для student_id %d, task_id %d с seminarist_id %d", workID, req.StudentId, req.TaskId, seminaristID)
	}

	// Фиксация транзакции
	if err = tx.Commit(); err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		return &pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}

	return &pb.CreateWorkResponse{WorkId: int32(workID)}, nil
}

func (s *server) DownloadAssignmentFile(ctx context.Context, req *pb.DownloadAssignmentFileRequest) (*pb.DownloadAssignmentFileResponse, error) {
	if ctx.Err() != nil {
		return &pb.DownloadAssignmentFileResponse{Error: "Request canceled"}, nil
	}

	// Генерация уникального ключа (например, с текущей датой и временем)
	bucket := "your-bucket-name" // Замените на имя вашего бакета
	key := fmt.Sprintf("works/%d/%s-%s", req.WorkId, time.Now().Format("20060102-150405"), req.FileName)

	// Загрузка файла в S3
	uploader := manager.NewUploader(S3Client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(req.Content),
	})
	if err != nil {
		log.Printf("Ошибка загрузки файла для work_id %d: %v", req.WorkId, err)
		return &pb.DownloadAssignmentFileResponse{Error: fmt.Sprintf("Ошибка загрузки: %v", err)}, nil
	}

	// Обновление content_url в базе данных
	_, err = s.db.ExecContext(ctx, `
		UPDATE student_works 
		SET content_url = $1 
		WHERE id = $2`,
		key, req.WorkId)
	if err != nil {
		log.Printf("Ошибка обновления content_url для work_id %d: %v", req.WorkId, err)
		return &pb.DownloadAssignmentFileResponse{Error: "Ошибка обновления базы данных"}, nil
	}

	// Генерация временной ссылки (presigned URL)
	presigner := s3.NewPresignClient(S3Client)
	presignResult, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour // Ссылка действительна 24 часа
	})
	if err != nil {
		log.Printf("Ошибка генерации presigned URL для work_id %d: %v", req.WorkId, err)
		return &pb.DownloadAssignmentFileResponse{Error: "Ошибка генерации ссылки"}, nil
	}

	log.Printf("Файл для work_id %d загружен в S3: %s", req.WorkId, key)
	return &pb.DownloadAssignmentFileResponse{ContentUrl: presignResult.URL}, nil
}

func (s *server) GetAssignmentFileURL(ctx context.Context, req *pb.GetAssignmentFileURLRequest) (*pb.GetAssignmentFileURLResponse, error) {
	if ctx.Err() != nil {
		return &pb.GetAssignmentFileURLResponse{Error: "Request canceled"}, nil
	}

	var contentURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT content_url 
		FROM student_works 
		WHERE id = $1`,
		req.WorkId).Scan(&contentURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetAssignmentFileURLResponse{Error: "Работа не найдена"}, nil
		}
		log.Printf("Ошибка получения content_url для work_id %d: %v", req.WorkId, err)
		return &pb.GetAssignmentFileURLResponse{Error: "Ошибка сервера"}, nil
	}

	if contentURL == "" {
		return &pb.GetAssignmentFileURLResponse{Error: "Файл не загружен"}, nil
	}

	// Генерация presigned URL
	bucket := "your-bucket-name" // Замените на имя вашего бакета
	presigner := s3.NewPresignClient(S3Client)
	presignResult, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &contentURL,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour
	})
	if err != nil {
		log.Printf("Ошибка генерации presigned URL для work_id %d: %v", req.WorkId, err)
		return &pb.GetAssignmentFileURLResponse{Error: "Ошибка генерации ссылки"}, nil
	}

	log.Printf("Сгенерирована ссылка для work_id %d: %s", req.WorkId, presignResult.URL)
	return &pb.GetAssignmentFileURLResponse{Url: presignResult.URL}, nil
}

func main() {

	initS3Client()

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
	pb.RegisterWorkAssignmentServiceServer(s, &server{db: db})

	log.Println("WorkAssignmentService starting on :50054")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
