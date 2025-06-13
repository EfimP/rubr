package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "rubr/proto/rubric"
	"strconv"
)

type server struct {
	pb.UnimplementedRubricServiceServer
	db *sql.DB
}

func (s *server) CreateNewBlockingCriteria(ctx context.Context, req *pb.CreateNewBlockingCriteriaRequest) (*pb.CreateNewBlockingCriteriaResponse, error) {
	var groupID int32
	query := `SELECT id FROM criteria_groups WHERE task_id = $1 AND group_name = 'blocking_criterias' AND block_flag = true`
	err := s.db.QueryRowContext(ctx, query, req.TaskId).Scan(&groupID)
	if err == sql.ErrNoRows {
		insertGroup := `INSERT INTO criteria_groups (task_id, group_name, block_flag) VALUES ($1, 'blocking_criterias', true) RETURNING id`
		err = s.db.QueryRowContext(ctx, insertGroup, req.TaskId).Scan(&groupID)
		if err != nil {
			return &pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
		}
	} else if err != nil {
		return &pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
	}

	insertCriteria := `INSERT INTO criteria (name, description, comment_for_blocking_criteria, final_mark_for_blocking_criteria, criteria_group_id, weight) 
		VALUES ($1, $2, $3, $4, $5, 0) RETURNING id`
	var criteriaID int
	err = s.db.QueryRowContext(ctx, insertCriteria, req.Name, req.Description, req.Comment, req.FinalMark, groupID).Scan(&criteriaID)
	if err != nil {
		return &pb.CreateNewBlockingCriteriaResponse{Error: err.Error()}, nil
	}
	return &pb.CreateNewBlockingCriteriaResponse{CriteriaGroupId: groupID}, nil
}

func (s *server) CreateNewCriteriaGroup(ctx context.Context, req *pb.CreateNewCriteriaGroupRequest) (*pb.CreateNewCriteriaGroupResponse, error) {
	query := `INSERT INTO criteria_groups (task_id, group_name) VALUES ($1, $2) RETURNING id`
	var groupID int32
	err := s.db.QueryRowContext(ctx, query, req.TaskId, req.GroupName).Scan(&groupID)
	if err != nil {
		return &pb.CreateNewCriteriaGroupResponse{Error: err.Error()}, nil
	}
	return &pb.CreateNewCriteriaGroupResponse{CriteriaGroupId: groupID}, nil
}

func (s *server) CreateNewMainCriteria(ctx context.Context, req *pb.CreateNewMainCriteriaRequest) (*pb.CreateNewMainCriteriaResponse, error) {
	query := `INSERT INTO criteria (name, criteria_group_id, weight) VALUES ($1, $2, 0) RETURNING id`
	var criteriaID int32
	err := s.db.QueryRowContext(ctx, query, req.Name, req.CriteriaGroupId).Scan(&criteriaID)
	if err != nil {
		return &pb.CreateNewMainCriteriaResponse{Error: err.Error()}, nil
	}
	return &pb.CreateNewMainCriteriaResponse{CriteriaId: criteriaID}, nil
}

func (s *server) CreateCriteriaDescription(ctx context.Context, req *pb.CreateCriteriaDescriptionRequest) (*pb.CreateCriteriaDescriptionResponse, error) {
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
		return &pb.CreateCriteriaDescriptionResponse{Error: "invalid mark value"}, nil
	}
	query := `UPDATE criteria SET ` + column + ` = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.Comment, req.CriteriaId)
	if err != nil {
		return &pb.CreateCriteriaDescriptionResponse{Error: err.Error()}, nil
	}
	return &pb.CreateCriteriaDescriptionResponse{Success: true}, nil
}

func (s *server) SetCriteriaWeight(ctx context.Context, req *pb.SetCriteriaWeightRequest) (*pb.SetCriteriaWeightResponse, error) {
	query := `UPDATE criteria SET weight = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, req.Weight, req.CriteriaId)
	if err != nil {
		return &pb.SetCriteriaWeightResponse{Error: err.Error()}, nil
	}
	return &pb.SetCriteriaWeightResponse{Success: true}, nil
}

func (s *server) LoadTaskBlockingCriterias(ctx context.Context, req *pb.LoadTaskBlockingCriteriasRequest) (*pb.LoadTaskBlockingCriteriasResponse, error) {
	query := `
		SELECT c.id, c.name, c.description, c.comment_for_blocking_criteria, c.final_mark_for_blocking_criteria
		FROM criteria c
		JOIN criteria_groups cg ON c.criteria_group_id = cg.id
		WHERE cg.task_id = $1 AND cg.group_name = 'blocking_criterias' AND cg.block_flag = true`
	rows, err := s.db.QueryContext(ctx, query, req.TaskId)
	if err != nil {
		return &pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var criteriaList []*pb.BlockingCriteria
	for rows.Next() {
		var crit pb.BlockingCriteria
		if err := rows.Scan(&crit.Id, &crit.Name, &crit.Description, &crit.Comment, &crit.FinalMark); err != nil {
			return &pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
		}
		criteriaList = append(criteriaList, &crit)
	}
	if err := rows.Err(); err != nil {
		return &pb.LoadTaskBlockingCriteriasResponse{Error: err.Error()}, nil
	}
	return &pb.LoadTaskBlockingCriteriasResponse{Criteria: criteriaList}, nil
}

func (s *server) LoadTaskMainCriterias(ctx context.Context, req *pb.LoadTaskMainCriteriasRequest) (*pb.LoadTaskMainCriteriasResponse, error) {
	// First, get all criteria groups for the task (excluding blocking criteria)
	queryGroups := `SELECT id, group_name FROM criteria_groups WHERE task_id = $1 AND block_flag = false`
	rows, err := s.db.QueryContext(ctx, queryGroups, req.TaskId)
	if err != nil {
		return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
	}
	defer rows.Close()

	var groups []*pb.CriteriaGroup
	for rows.Next() {
		var group pb.CriteriaGroup
		if err := rows.Scan(&group.Id, &group.GroupName); err != nil {
			return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		// Get criteria for this group
		queryCriteria := `
			SELECT id, name, weight, comment_000, comment_025, comment_050, comment_075, comment_100
			FROM criteria
			WHERE criteria_group_id = $1`
		critRows, err := s.db.QueryContext(ctx, queryCriteria, group.Id)
		if err != nil {
			return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		defer critRows.Close()

		for critRows.Next() {
			var crit pb.MainCriteria
			if err := critRows.Scan(&crit.Id, &crit.Name, &crit.Weight, &crit.Comment_000, &crit.Comment_025, &crit.Comment_050, &crit.Comment_075, &crit.Comment_100); err != nil {
				return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
			}
			group.Criteria = append(group.Criteria, &crit)
		}
		if err := critRows.Err(); err != nil {
			return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
		}
		groups = append(groups, &group)
	}
	if err := rows.Err(); err != nil {
		return &pb.LoadTaskMainCriteriasResponse{Error: err.Error()}, nil
	}
	return &pb.LoadTaskMainCriteriasResponse{Groups: groups}, nil
}

func main() {

	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	// конвертируем порт из строки в число, чтобы работал sql.open
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT value: %v", err)
	}

	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	log.Printf("Trying to connect to: %s", connStr) // Для отладки
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Настройка сервера gRPC (остальной код остается без изменений)
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRubricServiceServer(s, &server{db: db})

	log.Println("RubricService starting on :50055")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
