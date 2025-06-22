package workassignmentservice

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"os"
	"path/filepath"
	Pb "rubr/proto/workassignment"
	"time"
)

var myBucket = "fa9d45a5ad42-flexible-kenji"
var accessKey = "UNQCCBCI5X4I8IHAI9XF"
var accessSecret = "KddoCXMG5LHQvrO6GSB5UXFRrxP7rqtbuA1JyMm1"

var S3Client *s3.Client

func InitS3Client() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ru1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, accessSecret, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           "https://s3.ru1.storage.beget.cloud",
					SigningRegion: "ru1",
				}, nil
			},
		)),
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации S3 клиента: %v, Config: %+v", err, cfg)
	}

	S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	log.Printf("S3 клиент инициализирован с регионом: %s и endpoint: %s", cfg.Region, "https://s3.ru1.storage.beget.cloud")
}

func (s *Server) GenerateUploadURL(ctx context.Context, req *Pb.GenerateUploadURLRequest) (*Pb.GenerateUploadURLResponse, error) {
	// Генерация уникального ключа для файла в S3
	key := fmt.Sprintf("works/%d/%s", req.WorkId, req.FileName)

	// Создание pre-signed URL
	presigner := s3.NewPresignClient(S3Client)
	presignReq, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return &Pb.GenerateUploadURLResponse{
			Error: fmt.Sprintf("Ошибка генерации URL: %v", err),
		}, nil
	}

	// Сохранение content_url в базе данных
	_, err = s.Db.ExecContext(ctx, `
		UPDATE student_works 
		SET content_url = $1 
		WHERE id = $2`,
		key, req.WorkId)
	if err != nil {
		log.Printf("Ошибка обновления content_url для work_id %d: %v", req.WorkId, err)
		return &Pb.GenerateUploadURLResponse{
			Error: "Ошибка обновления базы данных",
		}, nil
	}

	return &Pb.GenerateUploadURLResponse{
		Url: presignReq.URL,
	}, nil
}

func (s *Server) DownloadAssignmentFile(ctx context.Context, req *Pb.DownloadAssignmentFileRequest) (*Pb.DownloadAssignmentFileResponse, error) {
	if ctx.Err() != nil {
		return &Pb.DownloadAssignmentFileResponse{Error: "Request canceled"}, nil
	}

	// Открытие файла
	baseDir := "C:/Users/User/Documents"
	fullPath := filepath.Join(baseDir, req.FileName)
	f, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Ошибка открытия файла %q для work_id %d: %v", req.FileName, req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: fmt.Sprintf("Ошибка открытия файла: %v", err)}, nil
	}
	defer f.Close()

	// Генерация уникального ключа
	key := fmt.Sprintf("works/%d/%s-%s", req.WorkId, time.Now().Format("20060102-150405"), req.FileName)

	// Вычисление SHA-256 хеша (опционально, для целостности)
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		log.Printf("Ошибка вычисления хеша для work_id %d: %v", req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: "Ошибка обработки данных"}, nil
	}
	contentSha256 := hex.EncodeToString(hash.Sum(nil))
	log.Printf("Вычисленный хеш для work_id %d: %s", req.WorkId, contentSha256)

	// Перемотка файла в начало
	if _, err := f.Seek(0, 0); err != nil {
		log.Printf("Ошибка перемотки файла для work_id %d: %v", req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: "Ошибка подготовки файла"}, nil
	}

	// Настройка uploader
	uploader := manager.NewUploader(S3Client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5MB части
		u.Concurrency = 2            // Количество параллельных загрузок
	})

	// Загрузка файла в S3
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(myBucket),
		Key:          aws.String(key),
		Body:         f,
		CacheControl: aws.String("no-cache"), // Установка Cache-Control: no-cache
	})
	if err != nil {
		log.Printf("Ошибка загрузки файла для work_id %d: %v", req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: fmt.Sprintf("Ошибка загрузки: %v", err)}, nil
	}

	// Обновление content_url в базе данных
	_, err = s.Db.ExecContext(ctx, `
		UPDATE student_works 
		SET content_url = $1 
		WHERE id = $2`,
		key, req.WorkId)
	if err != nil {
		log.Printf("Ошибка обновления content_url для work_id %d: %v", req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: "Ошибка обновления базы данных"}, nil
	}

	// Генерация presigned URL
	presigner := s3.NewPresignClient(S3Client)
	presignResult, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour
	})
	if err != nil {
		log.Printf("Ошибка генерации presigned URL для work_id %d: %v", req.WorkId, err)
		return &Pb.DownloadAssignmentFileResponse{Error: "Ошибка генерации ссылки"}, nil
	}

	log.Printf("Файл для work_id %d загружен в S3: %s", req.WorkId, key)
	return &Pb.DownloadAssignmentFileResponse{ContentUrl: presignResult.URL}, nil
}

func (s *Server) GetAssignmentFileURL(ctx context.Context, req *Pb.GetAssignmentFileURLRequest) (*Pb.GetAssignmentFileURLResponse, error) {
	if ctx.Err() != nil {
		return &Pb.GetAssignmentFileURLResponse{Error: "Request canceled"}, nil
	}

	var contentURL string
	err := s.Db.QueryRowContext(ctx, `
		SELECT content_url 
		FROM student_works 
		WHERE id = $1`,
		req.WorkId).Scan(&contentURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return &Pb.GetAssignmentFileURLResponse{Error: "Работа не найдена"}, nil
		}
		log.Printf("Ошибка получения content_url для work_id %d: %v", req.WorkId, err)
		return &Pb.GetAssignmentFileURLResponse{Error: "Ошибка сервера"}, nil
	}

	if contentURL == "" {
		return &Pb.GetAssignmentFileURLResponse{Error: "Файл не загружен"}, nil
	}

	// Генерация presigned URL
	bucket := "fa9d45a5ad42-flexible-kenji" // Замените на имя вашего бакета
	presigner := s3.NewPresignClient(S3Client)
	presignResult, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &contentURL,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour
	})
	if err != nil {
		log.Printf("Ошибка генерации presigned URL для work_id %d: %v", req.WorkId, err)
		return &Pb.GetAssignmentFileURLResponse{Error: "Ошибка генерации ссылки"}, nil
	}

	log.Printf("Сгенерирована ссылка для work_id %d: %s", req.WorkId, presignResult.URL)
	return &Pb.GetAssignmentFileURLResponse{Url: presignResult.URL}, nil
}

func (s *Server) UploadAssignmentFile(ctx context.Context, req *Pb.UploadAssignmentFileRequest) (*Pb.UploadAssignmentFileResponse, error) {
	query := `UPDATE student_works SET content_url = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.FilePath, req.WorkId)
	if err != nil {
		return &Pb.UploadAssignmentFileResponse{Error: err.Error()}, nil
	}
	return &Pb.UploadAssignmentFileResponse{Success: true}, nil
}
