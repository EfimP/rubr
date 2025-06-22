package workassignmentservice

import (
	"context"
	"database/sql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	Pb "rubr/proto/workassignment"
	"time"
)

func (s *Server) GetWorksForAssistant(ctx context.Context, req *Pb.GetWorksForAssistantRequest) (*Pb.GetWorksForAssistantResponse, error) {
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
	rows, err := s.Db.QueryContext(ctx, query, assistantID)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &Pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}
	defer rows.Close()

	var works []*Pb.WorkAssignment
	for rows.Next() {
		var work Pb.WorkAssignment
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
			return &Pb.GetWorksForAssistantResponse{Error: "Ошибка обработки данных"}, nil
		}
		works = append(works, &work)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка после обработки строк: %v", err)
		return &Pb.GetWorksForAssistantResponse{Error: "Ошибка сервера"}, nil
	}

	return &Pb.GetWorksForAssistantResponse{Works: works}, nil
}

func (s *Server) GetWorkDetails(ctx context.Context, req *Pb.GetWorkDetailsRequest) (*Pb.GetWorkDetailsResponse, error) {
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
	var work Pb.GetWorkDetailsResponse
	err := s.Db.QueryRowContext(ctx, query, workID).Scan(
		&work.WorkId,
		&work.TaskTitle,
		&work.TaskDescription,
		&work.TaskDeadline,
		&work.CreatedAt,
		&work.Status,
		&work.ContentUrl,
	)
	if err == sql.ErrNoRows {
		return &Pb.GetWorkDetailsResponse{Error: "Работа не найдена"}, nil
	}
	if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return &Pb.GetWorkDetailsResponse{Error: "Ошибка сервера"}, nil
	}

	return &work, nil
}

func (s *Server) GetTaskDetails(ctx context.Context, req *Pb.GetTaskDetailsRequest) (*Pb.GetTaskDetailsResponse, error) {
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

	rows, err := s.Db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		log.Printf("Ошибка запроса к базе данных для task_id %d: %v", req.TaskId, err)
		return nil, status.Errorf(codes.Internal, "Ошибка сервера: %v", err)
	}
	defer rows.Close()

	var response Pb.GetTaskDetailsResponse
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
		cg := &Pb.CriteriaGroup{
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

func (s *Server) CheckExistingWork(ctx context.Context, req *Pb.CheckExistingWorkRequest) (*Pb.CheckExistingWorkResponse, error) {
	if ctx.Err() != nil {
		return &Pb.CheckExistingWorkResponse{Error: "Request canceled"}, nil
	}

	var exists bool
	var workID int64
	err := s.Db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM student_works WHERE student_id = $1 AND task_id = $2),
               id
        FROM student_works 
        WHERE student_id = $1 AND task_id = $2`,
		req.StudentId, req.TaskId).Scan(&exists, &workID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Ошибка проверки работы для student_id %d и task_id %d: %v", req.StudentId, req.TaskId, err)
		return &Pb.CheckExistingWorkResponse{Error: "Ошибка сервера"}, nil
	}

	if !exists {
		return &Pb.CheckExistingWorkResponse{Exists: false, StudentId: req.StudentId}, nil
	}

	return &Pb.CheckExistingWorkResponse{
		Exists:    true,
		WorkId:    int32(workID),
		StudentId: req.StudentId,
	}, nil
}

func (s *Server) CheckStudentWorkExists(ctx context.Context, req *Pb.CheckStudentWorkExistsRequest) (*Pb.CheckStudentWorkExistsResponse, error) {
	if ctx.Err() != nil {
		return &Pb.CheckStudentWorkExistsResponse{Exists: false, Error: "Request canceled"}, nil
	}

	// Начало транзакции
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		return &Pb.CheckStudentWorkExistsResponse{Exists: false, Error: "Ошибка сервера"}, nil
	}
	defer tx.Rollback()

	var workID int64
	err = tx.QueryRowContext(ctx, `
        SELECT id FROM student_works WHERE student_id = $1 AND task_id = $2 FOR UPDATE`,
		req.UserId, req.TaskId).Scan(&workID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Работа для student_id %d и task_id %d не найдена", req.UserId, req.TaskId)
			return &Pb.CheckStudentWorkExistsResponse{Exists: false}, nil
		}
		log.Printf("Ошибка проверки работы для student_id %d и task_id %d: %v", req.UserId, req.TaskId, err)
		return &Pb.CheckStudentWorkExistsResponse{Exists: false, Error: "Ошибка сервера"}, nil
	}

	// Фиксация транзакции
	if err = tx.Commit(); err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		return &Pb.CheckStudentWorkExistsResponse{Exists: false, Error: "Ошибка сервера"}, nil
	}

	log.Printf("Работа с ID %d для student_id %d и task_id %d успешно найдена", workID, req.UserId, req.TaskId)
	return &Pb.CheckStudentWorkExistsResponse{Exists: true, WorkId: int32(workID)}, nil
}
