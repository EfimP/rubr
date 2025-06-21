package rubricservice

import (
	"context"
	"database/sql"
	"fmt"
	Pb "rubr/proto/rubric"
)

func (s *Server) UpdateCriterionWeight(ctx context.Context, req *Pb.UpdateCriterionWeightRequest) (*Pb.UpdateCriterionWeightResponse, error) {
	query := `UPDATE criteria SET weight = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Weight, req.CriterionId)
	if err != nil {
		return &Pb.UpdateCriterionWeightResponse{Error: err.Error()}, nil
	}
	return &Pb.UpdateCriterionWeightResponse{Success: true}, nil
}

func (s *Server) UpdateCriterionComment(ctx context.Context, req *Pb.UpdateCriterionCommentRequest) (*Pb.UpdateCriterionCommentResponse, error) {
	var column string
	switch req.Mark {
	case "000":
		column = "comment_000"
	case "025":
		column = "comment_025"
	case "050":
		column = "comment_050"
	case "075":
		column = "comment_075"
	case "100":
		column = "comment_100"
	default:
		return &Pb.UpdateCriterionCommentResponse{Error: "invalid mark value"}, nil
	}
	query := fmt.Sprintf("UPDATE criteria SET %s = $1 WHERE id = $2", column)
	_, err := s.Db.ExecContext(ctx, query, req.Comment, req.CriterionId)
	if err != nil {
		return &Pb.UpdateCriterionCommentResponse{Error: err.Error()}, nil
	}
	return &Pb.UpdateCriterionCommentResponse{Success: true}, nil
}

func (s *Server) CreateCriteriaGroup(ctx context.Context, req *Pb.CreateCriteriaGroupRequest) (*Pb.CreateCriteriaGroupResponse, error) {
	query := `INSERT INTO criteria_groups (task_id, group_name, block_flag) VALUES ($1, $2, false) RETURNING id`
	var groupID int32
	err := s.Db.QueryRowContext(ctx, query, req.TaskId, req.GroupName).Scan(&groupID)
	if err != nil {
		return &Pb.CreateCriteriaGroupResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateCriteriaGroupResponse{GroupId: groupID}, nil
}

func (s *Server) CreateCriterion(ctx context.Context, req *Pb.CreateCriterionRequest) (*Pb.CreateCriterionResponse, error) {
	query := `INSERT INTO criteria (name, criteria_group_id, weight) VALUES ($1, $2, 0) RETURNING id`
	var criterionID int32
	err := s.Db.QueryRowContext(ctx, query, req.Name, req.GroupId).Scan(&criterionID)
	if err != nil {
		return &Pb.CreateCriterionResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateCriterionResponse{CriterionId: criterionID}, nil
}

func (s *Server) CreateNewBlockingCriteria(ctx context.Context, req *Pb.CreateNewBlockingCriteriaRequest) (*Pb.CreateNewBlockingCriteriaResponse, error) {
	var groupID int32
	query := `SELECT id FROM criteria_groups WHERE task_id = $1 AND group_name = 'blocking_criterias' AND block_flag = true`
	err := s.Db.QueryRowContext(ctx, query, req.TaskId).Scan(&groupID)
	if err == sql.ErrNoRows {
		insertGroup := `INSERT INTO criteria_groups (task_id, group_name, block_flag) VALUES ($1, 'blocking_criterias', true) RETURNING id`
		err = s.Db.QueryRowContext(ctx, insertGroup, req.TaskId).Scan(&groupID)
		if err != nil {
			return &Pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
		}
	} else if err != nil {
		return &Pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
	}

	insertCriteria := `INSERT INTO criteria (name, description, comment_for_blocking_criteria, final_mark_for_blocking_criteria, criteria_group_id, weight) 
		VALUES ($1, $2, $3, $4, $5, 0) RETURNING id`
	var criteriaID int
	err = s.Db.QueryRowContext(ctx, insertCriteria, req.Name, req.Description, req.Comment, req.FinalMark, groupID).Scan(&criteriaID)
	if err != nil {
		return &Pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateNewBlockingCriteriaResponse{CriteriaGroupId: groupID}, nil
}

func (s *Server) CreateNewCriteriaGroup(ctx context.Context, req *Pb.CreateNewCriteriaGroupRequest) (*Pb.CreateNewCriteriaGroupResponse, error) {
	query := `INSERT INTO criteria_groups (task_id, group_name) VALUES ($1, $2) RETURNING id`
	var groupID int32
	err := s.Db.QueryRowContext(ctx, query, req.TaskId, req.GroupName).Scan(&groupID)
	if err != nil {
		return &Pb.CreateNewCriteriaGroupResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateNewCriteriaGroupResponse{CriteriaGroupId: groupID}, nil
}

func (s *Server) CreateNewMainCriteria(ctx context.Context, req *Pb.CreateNewMainCriteriaRequest) (*Pb.CreateNewMainCriteriaResponse, error) {
	query := `INSERT INTO criteria (name, criteria_group_id, weight) VALUES ($1, $2, 0) RETURNING id`
	var criteriaID int32
	err := s.Db.QueryRowContext(ctx, query, req.Name, req.CriteriaGroupId).Scan(&criteriaID)
	if err != nil {
		return &Pb.CreateNewMainCriteriaResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateNewMainCriteriaResponse{CriteriaId: criteriaID}, nil
}

func (s *Server) CreateCriteriaDescription(ctx context.Context, req *Pb.CreateCriteriaDescriptionRequest) (*Pb.CreateCriteriaDescriptionResponse, error) {
	var column string
	switch req.Mark {
	case "000":
		column = "comment_000"
	case "025":
		column = "comment_025"
	case "050":
		column = "comment_050"
	case "075":
		column = "comment_075"
	case "100":
		column = "comment_100"
	default:
		return &Pb.CreateCriteriaDescriptionResponse{Error: "invalid mark value"}, nil
	}
	query := `UPDATE criteria SET ` + column + ` = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Comment, req.CriteriaId)
	if err != nil {
		return &Pb.CreateCriteriaDescriptionResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateCriteriaDescriptionResponse{Success: true}, nil
}

func (s *Server) SetCriteriaWeight(ctx context.Context, req *Pb.SetCriteriaWeightRequest) (*Pb.SetCriteriaWeightResponse, error) {
	query := `UPDATE criteria SET weight = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Weight, req.CriteriaId)
	if err != nil {
		return &Pb.SetCriteriaWeightResponse{Error: err.Error()}, nil
	}
	return &Pb.SetCriteriaWeightResponse{Success: true}, nil
}
