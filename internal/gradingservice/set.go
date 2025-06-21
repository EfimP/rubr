package gradingservice

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	Pb "rubr/proto/grade"
)

func (s *Server) SetBlockingCriteriaMark(ctx context.Context, req *Pb.SetBlockingCriteriaMarkRequest) (*Pb.SetBlockingCriteriaMarkResponse, error) {
	query := `
        INSERT INTO student_criteria_marks (student_work_id, criteria_id, mark, comment)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (student_work_id, criteria_id) DO UPDATE 
        SET mark = EXCLUDED.mark, comment = EXCLUDED.comment`
	_, err := s.Db.ExecContext(ctx, query, req.WorkId, req.CriterionId, req.Mark, req.Comment)
	if err != nil {
		return &Pb.SetBlockingCriteriaMarkResponse{Error: err.Error()}, nil
	}
	return &Pb.SetBlockingCriteriaMarkResponse{}, nil
}

func (s *Server) SetMainCriteriaMark(ctx context.Context, req *Pb.SetMainCriteriaMarkRequest) (*Pb.SetMainCriteriaMarkResponse, error) {
	query := `
        INSERT INTO student_criteria_marks (student_work_id, criteria_id, mark, comment)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (student_work_id, criteria_id) DO UPDATE 
        SET mark = EXCLUDED.mark, comment = EXCLUDED.comment`
	_, err := s.Db.ExecContext(ctx, query, req.WorkId, req.CriterionId, req.Mark, req.Comment)
	if err != nil {
		return &Pb.SetMainCriteriaMarkResponse{Error: err.Error()}, nil
	}
	return &Pb.SetMainCriteriaMarkResponse{}, nil
}

func (s *Server) UpdateWorkStatus(ctx context.Context, req *Pb.UpdateWorkStatusRequest) (*Pb.UpdateWorkStatusResponse, error) {
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
	err := s.Db.QueryRowContext(ctx, query, req.Status, req.WorkId).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Работа с id %d не найдена", req.WorkId)
			return &Pb.UpdateWorkStatusResponse{Error: fmt.Sprintf("работа с id %d не найдена", req.WorkId)}, nil
		}
		log.Printf("Ошибка обновления статуса для work_id %d: %v", req.WorkId, err)
		return nil, status.Errorf(codes.Internal, "ошибка базы данных: %v", err)
	}

	log.Printf("Статус работы %d успешно обновлен на %s", req.WorkId, req.Status)
	return &Pb.UpdateWorkStatusResponse{}, nil
}
