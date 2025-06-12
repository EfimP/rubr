package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

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

func (r *Repository) ListGroups(ctx context.Context) ([]*pb.Group, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description FROM student_groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*pb.Group
	for rows.Next() {
		var id int32
		var name, description string
		if err := rows.Scan(&id, &name, &description); err != nil {
			return nil, err
		}
		groups = append(groups, &pb.Group{Id: id, Name: name, Description: description})
	}
	return groups, nil
}

func (r *Repository) ManageGroupEntity(ctx context.Context, groupID int32, name, description, action string) (int32, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var newGroupID int32
	if action == "create" {
		if name == "" {
			return 0, fmt.Errorf("name is required for creating a group")
		}
		query := "INSERT INTO student_groups (name, description) VALUES ($1, $2) RETURNING id"
		err = tx.QueryRowContext(ctx, query, name, description).Scan(&newGroupID)
		if err != nil {
			return 0, err
		}
	} else if action == "delete" {
		if groupID <= 0 {
			return 0, fmt.Errorf("invalid group ID for deletion")
		}
		query := "DELETE FROM student_groups WHERE id = $1"
		result, err := tx.ExecContext(ctx, query, groupID)
		if err != nil {
			return 0, err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return 0, fmt.Errorf("group with ID %d not found", groupID)
		}
		newGroupID = 0 // Нет нового ID при удалении
	} else {
		return 0, fmt.Errorf("invalid action, must be 'create' or 'delete'")
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newGroupID, nil
}

type Service struct {
	repo *Repository
	pb.UnimplementedSuperAccServiceServer
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateUserRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	validRoles := map[string]bool{
		"student":      true,
		"assistant":    true,
		"seminarist":   true,
		"lecturer":     true,
		"superaccount": true,
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

func (s *Service) ListGroups(ctx context.Context, req *pb.ListGroupsRequest) (*pb.ListGroupsResponse, error) {
	groups, err := s.repo.ListGroups(ctx)
	if err != nil {
		return &pb.ListGroupsResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ListGroupsResponse{Success: true, Groups: groups}, nil
}

func (s *Service) ManageGroupEntity(ctx context.Context, req *pb.ManageGroupEntityRequest) (*pb.ManageGroupEntityResponse, error) {
	newGroupID, err := s.repo.ManageGroupEntity(ctx, req.GroupId, req.Name, req.Description, req.Action)
	if err != nil {
		return &pb.ManageGroupEntityResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ManageGroupEntityResponse{Message: "Group entity managed successfully", Success: true, GroupId: newGroupID}, nil
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

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
