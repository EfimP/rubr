package workservice

import (
	"context"
	Pb "rubr/proto/work"
)

func (s *Server) LoadTaskName(ctx context.Context, req *Pb.LoadTaskNameRequest) (*Pb.LoadTaskNameResponse, error) {
	query := `SELECT title FROM tasks WHERE id = $1`
	var title string
	err := s.Db.QueryRowContext(ctx, query, req.TaskId).Scan(&title)
	if err != nil {
		return &Pb.LoadTaskNameResponse{Error: err.Error()}, nil
	}
	return &Pb.LoadTaskNameResponse{Title: title}, nil
}

func (s *Server) LoadTaskDescription(ctx context.Context, req *Pb.LoadTaskDescriptionRequest) (*Pb.LoadTaskDescriptionResponse, error) {
	query := `SELECT description FROM tasks WHERE id = $1`
	var description string
	err := s.Db.QueryRowContext(ctx, query, req.TaskId).Scan(&description)
	if err != nil {
		return &Pb.LoadTaskDescriptionResponse{Error: err.Error()}, nil
	}
	return &Pb.LoadTaskDescriptionResponse{Description: description}, nil
}

func (s *Server) LoadTaskDeadline(ctx context.Context, req *Pb.LoadTaskDeadlineRequest) (*Pb.LoadTaskDeadlineResponse, error) {
	query := `SELECT deadline FROM tasks WHERE id = $1`
	var deadline string
	err := s.Db.QueryRowContext(ctx, query, req.TaskId).Scan(&deadline)
	if err != nil {
		return &Pb.LoadTaskDeadlineResponse{Error: err.Error()}, nil
	}
	return &Pb.LoadTaskDeadlineResponse{Deadline: deadline}, nil
}
