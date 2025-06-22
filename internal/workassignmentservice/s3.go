package workassignmentservice

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
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

func (s *Server) GenerateDownloadURL(ctx context.Context, req *Pb.GenerateDownloadURLRequest) (*Pb.GenerateDownloadURLResponse, error) {
	// Получение content_url из базы данных
	var contentURL string
	err := s.Db.QueryRowContext(ctx, `
		SELECT content_url 
		FROM student_works 
		WHERE id = $1`,
		req.WorkId).Scan(&contentURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return &Pb.GenerateDownloadURLResponse{
				Error: "Работа не найдена",
			}, nil
		}
		log.Printf("Ошибка получения content_url для work_id %d: %v", req.WorkId, err)
		return &Pb.GenerateDownloadURLResponse{
			Error: "Ошибка сервера",
		}, nil
	}

	if contentURL == "" {
		return &Pb.GenerateDownloadURLResponse{
			Error: "Файл не загружен",
		}, nil
	}

	// Генерация pre-signed URL для скачивания
	presigner := s3.NewPresignClient(S3Client)
	presignResult, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(contentURL),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour
	})
	if err != nil {
		log.Printf("Ошибка генерации presigned URL для work_id %d: %v", req.WorkId, err)
		return &Pb.GenerateDownloadURLResponse{
			Error: "Ошибка генерации ссылки",
		}, nil
	}

	log.Printf("Сгенерирована ссылка для скачивания для work_id %d: %s", req.WorkId, presignResult.URL)
	return &Pb.GenerateDownloadURLResponse{
		Url: presignResult.URL,
	}, nil
}
