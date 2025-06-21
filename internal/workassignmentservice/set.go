package workassignmentservice

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	Pb "rubr/proto/workassignment"
)

func (s *Server) SubmitWork(ctx context.Context, req *Pb.SubmitWorkRequest) (*Pb.SubmitWorkResponse, error) {
	query := `UPDATE student_works SET status = 'submitted', content_url = $1, created_at = CURRENT_TIMESTAMP WHERE id = $2`
	result, err := s.Db.ExecContext(ctx, query, req.FilePath, req.WorkId)
	if err != nil {
		log.Printf("Ошибка обновления работы %d: %v", req.WorkId, err)
		return &Pb.SubmitWorkResponse{Error: err.Error()}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Printf("Работа %d не найдена или не обновлена", req.WorkId)
		return &Pb.SubmitWorkResponse{Error: "Работа не найдена"}, nil
	}
	log.Printf("Работа %d успешно сдана", req.WorkId)
	return &Pb.SubmitWorkResponse{Success: true}, nil
}

func (s *Server) CreateWork(ctx context.Context, req *Pb.CreateWorkRequest) (*Pb.CreateWorkResponse, error) {
	if ctx.Err() != nil {
		return &Pb.CreateWorkResponse{Error: "Request canceled"}, nil
	}

	// Начало транзакции
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	defer tx.Rollback()

	var studentExists, taskExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.StudentId).Scan(&studentExists)
	if err != nil {
		log.Printf("Ошибка проверки студента %d: %v", req.StudentId, err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	if !studentExists {
		return &Pb.CreateWorkResponse{Error: "Студент не найден"}, nil
	}

	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", req.TaskId).Scan(&taskExists)
	if err != nil {
		log.Printf("Ошибка проверки задания %d: %v", req.TaskId, err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}
	if !taskExists {
		return &Pb.CreateWorkResponse{Error: "Задание не найдено"}, nil
	}

	// Извлечение group_id из tasks
	var groupID int64
	err = tx.QueryRowContext(ctx, `
        SELECT group_id FROM tasks WHERE id = $1`, req.TaskId).Scan(&groupID)
	if err != nil {
		log.Printf("Ошибка получения group_id для task_id %d: %v", req.TaskId, err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}

	// Извлечение seminarist_id из groups_in_disciplines
	var seminaristID int64
	err = tx.QueryRowContext(ctx, `
        SELECT seminarist_id FROM groups_in_disciplines WHERE group_id = $1`, groupID).Scan(&seminaristID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Не найден seminarist_id для group_id %d", groupID)
			return &Pb.CreateWorkResponse{Error: "Семинарист не назначен для группы"}, nil
		}
		log.Printf("Ошибка получения seminarist_id для group_id %d: %v", groupID, err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
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
			return &Pb.CreateWorkResponse{Error: "Ошибка создания работы"}, nil
		}
		log.Printf("Создано новая работа с ID %d для student_id %d, task_id %d, seminarist_id %d", workID, req.StudentId, req.TaskId, seminaristID)
	} else if err != nil {
		log.Printf("Ошибка проверки существующей работы для student_id %d и task_id %d: %v", req.StudentId, req.TaskId, err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	} else {
		// Если работа существует, обновляем данные
		_, err = tx.ExecContext(ctx, `
            UPDATE student_works 
            SET status = 'submitted', seminarist_id = $2
            WHERE id = $1`,
			workID, seminaristID)
		if err != nil {
			log.Printf("Ошибка обновления работы %d для student_id %d и task_id %d: %v", workID, req.StudentId, req.TaskId, err)
			return &Pb.CreateWorkResponse{Error: fmt.Sprintf("Ошибка обновления работы: %v", err)}, nil
		}
		log.Printf("Обновлена работа с ID %d для student_id %d, task_id %d с seminarist_id %d", workID, req.StudentId, req.TaskId, seminaristID)
	}

	// Фиксация транзакции
	if err = tx.Commit(); err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		return &Pb.CreateWorkResponse{Error: "Ошибка сервера"}, nil
	}

	return &Pb.CreateWorkResponse{WorkId: int32(workID)}, nil
}
