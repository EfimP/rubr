package workservice

import (
	"context"
	Pb "rubr/proto/work"
)

func (s *Server) DeleteTask(ctx context.Context, req *Pb.DeleteTaskRequest) (*Pb.DeleteTaskResponse, error) {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := s.Db.ExecContext(ctx, query, req.TaskId)
	if err != nil {
		return &Pb.DeleteTaskResponse{Error: err.Error()}, nil
	}
	return &Pb.DeleteTaskResponse{Success: true}, nil
}
