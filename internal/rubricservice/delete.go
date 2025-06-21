package rubricservice

import (
	"context"
	"database/sql"
	Pb "rubr/proto/rubric"
)

func (s *Server) DeleteCriteriaGroup(ctx context.Context, req *Pb.DeleteCriteriaGroupRequest) (*Pb.DeleteCriteriaGroupResponse, error) {
	query := `DELETE FROM criteria_groups WHERE id = $1 AND block_flag = false`
	result, err := s.Db.ExecContext(ctx, query, req.GroupId)
	if err != nil {
		return &Pb.DeleteCriteriaGroupResponse{Error: err.Error()}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &Pb.DeleteCriteriaGroupResponse{Error: err.Error()}, nil
	}
	if rowsAffected == 0 {
		return &Pb.DeleteCriteriaGroupResponse{Error: "Группа с указанным ID не найдена или является блокирующей"}, nil
	}
	return &Pb.DeleteCriteriaGroupResponse{Success: true}, nil
}

func (s *Server) DeleteCriterion(ctx context.Context, req *Pb.DeleteCriterionRequest) (*Pb.DeleteCriterionResponse, error) {
	query := `DELETE FROM criteria WHERE id = $1`
	result, err := s.Db.ExecContext(ctx, query, req.CriterionId)
	if err != nil {
		return &Pb.DeleteCriterionResponse{Error: err.Error()}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &Pb.DeleteCriterionResponse{Error: err.Error()}, nil
	}
	if rowsAffected == 0 {
		return &Pb.DeleteCriterionResponse{Error: "Критерий с указанным ID не найден"}, nil
	}
	return &Pb.DeleteCriterionResponse{Success: true}, nil
}

func (s *Server) DeleteBlockingCriteria(ctx context.Context, req *Pb.DeleteBlockingCriteriaRequest) (*Pb.DeleteBlockingCriteriaResponse, error) {
	query := `DELETE FROM criteria WHERE id = $1`
	result, err := s.Db.ExecContext(ctx, query, req.CriteriaId)
	if err != nil {
		return &Pb.DeleteBlockingCriteriaResponse{Error: err.Error()}, nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &Pb.DeleteBlockingCriteriaResponse{Error: err.Error()}, nil
	}
	if rowsAffected == 0 {
		return &Pb.DeleteBlockingCriteriaResponse{Error: "Критерий с указанным ID не найден"}, nil
	}
	return &Pb.DeleteBlockingCriteriaResponse{Success: true}, nil
}

func (s *Server) DeleteTaskBlockingCriterias(ctx context.Context, req *Pb.DeleteTaskBlockingCriteriasRequest) (*Pb.DeleteTaskBlockingCriteriasResponse, error) {
	// Находим ID группы критериев с block_flag = true и group_name = 'blocking_criterias'
	var groupID int32
	queryGroup := `SELECT id FROM criteria_groups WHERE task_id = $1 AND group_name = 'blocking_criterias' AND block_flag = true`
	err := s.Db.QueryRowContext(ctx, queryGroup, req.TaskId).Scan(&groupID)
	if err == sql.ErrNoRows {
		return &Pb.DeleteTaskBlockingCriteriasResponse{Success: true}, nil
	}
	if err != nil {
		return &Pb.DeleteTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}

	// Удаляем все критерии, связанные с этой группой
	queryCriteria := `DELETE FROM criteria WHERE criteria_group_id = $1`
	_, err = s.Db.ExecContext(ctx, queryCriteria, groupID)
	if err != nil {
		return &Pb.DeleteTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}

	// Удаляем саму группу критериев
	queryGroupDelete := `DELETE FROM criteria_groups WHERE id = $1`
	_, err = s.Db.ExecContext(ctx, queryGroupDelete, groupID)
	if err != nil {
		return &Pb.DeleteTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}

	return &Pb.DeleteTaskBlockingCriteriasResponse{Success: true}, nil
}
