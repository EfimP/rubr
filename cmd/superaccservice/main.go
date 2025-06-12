package superaccservice

import (
	"context"
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	pb "rubr/proto/superacc"
)

type SuperAcc struct {
	ID        int
	CreatedAt int64
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID int, role string) error {
	query := "UPDATE users SET role = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, role, userID)
	return err
}

func (r *Repository) ManageGroup(ctx context.Context, groupID int, action string, userID int, role string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if action == "add" {
		query := "INSERT INTO users_in_groups (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
		_, err = tx.ExecContext(ctx, query, userID, groupID)
	} else if action == "remove" {
		query := "DELETE FROM users_in_groups WHERE user_id = $1 AND group_id = $2"
		_, err = tx.ExecContext(ctx, query, userID, groupID)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) ManageDiscipline(ctx context.Context, disciplineID, groupID, seminaristID, assistantID int) error {
	query := `
		UPDATE groups_in_disciplines 
		SET seminarist_id = $1, assistant_id = $2 
		WHERE discipline_id = $3 AND group_id = $4`
	_, err := r.db.ExecContext(ctx, query, seminaristID, assistantID, disciplineID, groupID)
	return err
}

type Service struct {
	repo *Repository
	pb.UnimplementedSuperAccServiceServer
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// new comment in new branch

func (s *Service) UpdateUserRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	validRoles := map[string]bool{
		"student":    true,
		"assistant":  true,
		"seminarist": true,
		"lecturer":   false,
	}

	if req.UserId <= 0 {
		return &pb.UpdateRoleResponse{Message: "invalid user ID", Success: false}, nil
	}
	if req.Role == "" {
		return &pb.UpdateRoleResponse{Message: "role cannot be empty", Success: false}, nil
	}
	if !validRoles[req.Role] {
		return &pb.UpdateRoleResponse{Message: "invalid role, must be one of: student, assistant, seminarist, lecturer, superaccount", Success: false}, nil
	}

	err := s.repo.UpdateUserRole(ctx, int(req.UserId), req.Role)
	if err != nil {
		return &pb.UpdateRoleResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.UpdateRoleResponse{Message: "Role updated successfully", Success: true}, nil
}

func (s *Service) ManageGroup(ctx context.Context, req *pb.ManageGroupRequest) (*pb.ManageGroupResponse, error) {
	if req.GroupId <= 0 || req.UserId <= 0 || req.Action == "" || req.Role == "" {
		return &pb.ManageGroupResponse{Message: "invalid input parameters", Success: false}, nil
	}
	if req.Action != "add" && req.Action != "remove" {
		return &pb.ManageGroupResponse{Message: "invalid action", Success: false}, nil
	}
	err := s.repo.ManageGroup(ctx, int(req.GroupId), req.Action, int(req.UserId), req.Role)
	if err != nil {
		return &pb.ManageGroupResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ManageGroupResponse{Message: "Group managed successfully", Success: true}, nil
}

func (s *Service) ManageDiscipline(ctx context.Context, req *pb.ManageDisciplineRequest) (*pb.ManageDisciplineResponse, error) {
	if req.DisciplineId <= 0 || req.GroupId <= 0 {
		return &pb.ManageDisciplineResponse{Message: "invalid discipline or group ID", Success: false}, nil
	}
	err := s.repo.ManageDiscipline(ctx, int(req.DisciplineId), int(req.GroupId), int(req.SeminaristId), int(req.AssistantId))
	if err != nil {
		return &pb.ManageDisciplineResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ManageDisciplineResponse{Message: "Discipline managed successfully", Success: true}, nil
}

func main() {
	connStr := "user=postgres password=postgres dbname=rubrlocal sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSuperAccServiceServer(s, svc)

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
