package workservice

import (
	"context"
	"fmt"
	"log"
	Pb "rubr/proto/work"
)

func (s *Server) CreateWork(ctx context.Context, req *Pb.CreateWorkRequest) (*Pb.CreateWorkResponse, error) {
	query := `INSERT INTO tasks (lector_id, group_id, title, description, deadline, discipline_id, content_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var taskID int32
	err := s.Db.QueryRowContext(ctx, query, req.LectorId, req.GroupId, req.Title, req.Description, req.Deadline, req.DisciplineId, req.ContentUrl).Scan(&taskID)
	if err != nil {
		return &Pb.CreateWorkResponse{Error: err.Error()}, nil
	}
	return &Pb.CreateWorkResponse{TaskId: taskID}, nil
}

func (s *Server) UpdateWork(ctx context.Context, req *Pb.UpdateWorkRequest) (*Pb.UpdateWorkResponse, error) {
	_, err := s.Db.ExecContext(ctx, `
		UPDATE student_works
		SET status = $1
		WHERE id = $2
	`, req.Status, req.WorkId)
	if err != nil {
		log.Printf("Failed to update work %d: %v", req.WorkId, err)
		return &Pb.UpdateWorkResponse{Error: err.Error()}, err
	}
	return &Pb.UpdateWorkResponse{}, nil
}

func (s *Server) AssignAssistantsToWorks(ctx context.Context, req *Pb.AssignAssistantsToWorksRequest) (*Pb.AssignAssistantsToWorksResponse, error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return &Pb.AssignAssistantsToWorksResponse{
			Error: fmt.Sprintf("Failed to begin transaction: %v", err),
		}, nil
	}
	defer tx.Rollback()

	query := `UPDATE student_works SET assistant_id = $1 WHERE id = $2`
	for _, assignment := range req.Assignments {
		_, err := tx.ExecContext(ctx, query, assignment.AssistantId, assignment.WorkId)
		if err != nil {
			log.Printf("Failed to update assistant_id for work %d: %v", assignment.WorkId, err)
			return &Pb.AssignAssistantsToWorksResponse{
				Error: fmt.Sprintf("Failed to update assistant_id for work %d: %v", assignment.WorkId, err),
			}, nil
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return &Pb.AssignAssistantsToWorksResponse{
			Error: fmt.Sprintf("Failed to commit transaction: %v", err),
		}, nil
	}

	return &Pb.AssignAssistantsToWorksResponse{
		Success: true,
	}, nil
}

func (s *Server) SetTaskTitle(ctx context.Context, req *Pb.SetTaskTitleRequest) (*Pb.SetTaskTitleResponse, error) {
	query := `UPDATE tasks SET title = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Title, req.TaskId)
	if err != nil {
		return &Pb.SetTaskTitleResponse{Error: err.Error()}, nil
	}
	return &Pb.SetTaskTitleResponse{Success: true}, nil
}

func (s *Server) SetTaskDescription(ctx context.Context, req *Pb.SetTaskDescriptionRequest) (*Pb.SetTaskDescriptionResponse, error) {
	query := `UPDATE tasks SET description = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Description, req.TaskId)
	if err != nil {
		return &Pb.SetTaskDescriptionResponse{Error: err.Error()}, nil
	}
	return &Pb.SetTaskDescriptionResponse{Success: true}, nil
}

func (s *Server) SetTaskDeadline(ctx context.Context, req *Pb.SetTaskDeadlineRequest) (*Pb.SetTaskDeadlineResponse, error) {
	query := `UPDATE tasks SET deadline = $1 WHERE id = $2`
	_, err := s.Db.ExecContext(ctx, query, req.Deadline, req.TaskId)
	if err != nil {
		return &Pb.SetTaskDeadlineResponse{Error: err.Error()}, nil
	}
	return &Pb.SetTaskDeadlineResponse{Success: true}, nil
}

func (s *Server) UpdateTaskGroupAndDiscipline(ctx context.Context, req *Pb.UpdateTaskGroupAndDisciplineRequest) (*Pb.UpdateTaskGroupAndDisciplineResponse, error) {
	query := `UPDATE tasks SET group_id = $1, discipline_id = $2 WHERE id = $3`
	_, err := s.Db.ExecContext(ctx, query, req.GroupId, req.DisciplineId, req.TaskId)
	if err != nil {
		return &Pb.UpdateTaskGroupAndDisciplineResponse{Error: err.Error()}, nil
	}
	return &Pb.UpdateTaskGroupAndDisciplineResponse{Success: true}, nil
}
