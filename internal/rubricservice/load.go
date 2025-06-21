package rubricservice

import (
	"context"
	Pb "rubr/proto/rubric"
)

func (s *Server) LoadTaskBlockingCriterias(ctx context.Context, req *Pb.LoadTaskBlockingCriteriasRequest) (*Pb.LoadTaskBlockingCriteriasResponse, error) {
	query := `
		SELECT c.id, c.name, c.description, c.comment_for_blocking_criteria, c.final_mark_for_blocking_criteria
		FROM criteria c
		JOIN criteria_groups cg ON c.criteria_group_id = cg.id
		WHERE cg.task_id = $1 AND cg.group_name = 'blocking_criterias' AND cg.block_flag = true`
	rows, err := s.Db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		return &Pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var criteriaList []*Pb.BlockingCriteria
	for rows.Next() {
		var crit Pb.BlockingCriteria
		if err := rows.Scan(&crit.Id, &crit.Name, &crit.Description, &crit.Comment, &crit.FinalMark); err != nil {
			return &Pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
		}
		criteriaList = append(criteriaList, &crit)
	}
	if err := rows.Err(); err != nil {
		return &Pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}
	return &Pb.LoadTaskBlockingCriteriasResponse{Criteria: criteriaList}, nil
}

func (s *Server) LoadTaskMainCriterias(ctx context.Context, req *Pb.LoadTaskMainCriteriasRequest) (*Pb.LoadTaskMainCriteriasResponse, error) {
	// First, get all criteria groups for the task (excluding blocking criteria)
	queryGroups := `SELECT id, group_name FROM criteria_groups WHERE task_id = $1 AND block_flag = false`
	rows, err := s.Db.QueryContext(ctx, queryGroups, req.TaskId)
	if err != nil {
		return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var groups []*Pb.CriteriaGroup
	for rows.Next() {
		var group Pb.CriteriaGroup
		if err := rows.Scan(&group.Id, &group.GroupName); err != nil {
			return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		// Get criteria for this group
		queryCriteria := `
			SELECT id, name, weight, comment_000, comment_025, comment_050, comment_075, comment_100
			FROM criteria
			WHERE criteria_group_id = $1`
		critRows, err := s.Db.QueryContext(ctx, queryCriteria, group.Id)
		if err != nil {
			return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		defer critRows.Close()

		for critRows.Next() {
			var crit Pb.MainCriteria
			if err := critRows.Scan(&crit.Id, &crit.Name, &crit.Weight, &crit.Comment_000, &crit.Comment_025, &crit.Comment_050, &crit.Comment_075, &crit.Comment_100); err != nil {
				return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
			}
			group.Criteria = append(group.Criteria, &crit)
		}
		if err := critRows.Err(); err != nil {
			return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		groups = append(groups, &group)
	}
	if err := rows.Err(); err != nil {
		return &Pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
	}
	return &Pb.LoadTaskMainCriteriasResponse{Groups: groups}, nil
}
