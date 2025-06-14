package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

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

type Service struct {
	repo *Repository
	pb.UnimplementedSuperAccServiceServer
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (r *Repository) AddUser(ctx context.Context, fio, email, group, status string) (int32, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if email == "" {
		return 0, fmt.Errorf("email is required")
	}

	// Проверяем, существует ли пользователь с таким email
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("user with email %s already exists", email)
	}

	// Разделяем FIO на name, surname, patronymic
	parts := strings.Fields(fio)
	name := parts[0]
	surname := parts[1]
	var patronymic string
	if len(parts) > 2 {
		patronymic = parts[2]
	}

	// Генерируем временный пароль
	defaultPassword := "temp123"

	query := "INSERT INTO users (name, surname, patronymic, email, password, role) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	var newUserID int32
	err = tx.QueryRowContext(ctx, query, name, surname, patronymic, email, defaultPassword, status).Scan(&newUserID)
	if err != nil {
		return 0, err
	}

	// Если указана группа, добавляем в users_in_groups
	if group != "" {
		var groupID int32
		err = tx.QueryRowContext(ctx, "SELECT id FROM student_groups WHERE name = $1", group).Scan(&groupID)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("group %s not found", group)
			}
			return 0, err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO users_in_groups (user_id, group_id) VALUES ($1, $2)", newUserID, groupID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newUserID, nil
}

func (s *Service) AddUser(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	if req.Email == "" {
		return &pb.AddUserResponse{Message: "email is required", Success: false}, nil
	}
	newUserID, err := s.repo.AddUser(ctx, req.Fio, req.Email, req.Group, req.Status)
	if err != nil {
		return &pb.AddUserResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.AddUserResponse{Message: "User added successfully", Success: true, UserId: newUserID}, nil
}

func (r *Repository) RemoveUser(ctx context.Context, email string) error {
	query := "DELETE FROM users WHERE email = $1"
	result, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("user with email %s not found", email)
	}
	return nil
}

func (s *Service) RemoveUser(ctx context.Context, req *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	if req.Email == "" {
		return &pb.RemoveUserResponse{Message: "email is required", Success: false}, nil
	}
	err := s.repo.RemoveUser(ctx, req.Email)
	if err != nil {
		return &pb.RemoveUserResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.RemoveUserResponse{Message: "User removed successfully", Success: true}, nil
}

func (r *Repository) GetGroupStaff(ctx context.Context, groupID int32) (*pb.GetGroupStaffResponse, error) {
	query := `
        SELECT seminarist_id, assistant_id 
        FROM groups_in_disciplines 
        WHERE group_id = $1 LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, groupID)

	var seminaristID, assistantID sql.NullInt32
	err := row.Scan(&seminaristID, &assistantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetGroupStaffResponse{Success: true, SeminaristId: 0, AssistantId: 0}, nil
		}
		return nil, err
	}

	return &pb.GetGroupStaffResponse{
		Success:      true,
		SeminaristId: seminaristID.Int32,
		AssistantId:  assistantID.Int32,
	}, nil
}

func (s *Service) GetGroupStaff(ctx context.Context, req *pb.GetGroupStaffRequest) (*pb.GetGroupStaffResponse, error) {
	if req.GroupId <= 0 {
		return &pb.GetGroupStaffResponse{Message: "invalid group ID", Success: false}, nil
	}
	resp, err := s.repo.GetGroupStaff(ctx, req.GroupId)
	if err != nil {
		return &pb.GetGroupStaffResponse{Message: err.Error(), Success: false}, err
	}
	return resp, nil
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID int, role string) error {
	query := "UPDATE users SET role = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, role, userID)
	return err
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

func (r *Repository) ListAllUsers(ctx context.Context) ([]*pb.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, surname, patronymic, email, role 
		FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var id int32
		var name, surname, patronymic, email, role string
		if err := rows.Scan(&id, &name, &surname, &patronymic, &email, &role); err != nil {
			return nil, err
		}
		fio := fmt.Sprintf("%s %s %s", name, surname, patronymic)
		if patronymic == "" {
			fio = fmt.Sprintf("%s %s", name, surname)
		}
		users = append(users, &pb.User{
			Id:     id,
			Fio:    fio,
			Email:  email,
			Group:  "", // Группа определяется через users_in_groups, если нужно
			Status: role,
		})
	}
	return users, nil
}

func (s *Service) ListAllUsers(ctx context.Context, req *pb.ListAllUsersRequest) (*pb.ListAllUsersResponse, error) {
	users, err := s.repo.ListAllUsers(ctx)
	if err != nil {
		return &pb.ListAllUsersResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ListAllUsersResponse{Success: true, Users: users}, nil
}

func (s *Service) ListUsersByGroup(ctx context.Context, req *pb.ListUsersByGroupRequest) (*pb.ListUsersByGroupResponse, error) {
	users, err := s.repo.ListUsersByGroup(ctx, req.GroupName)
	if err != nil {
		return &pb.ListUsersByGroupResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ListUsersByGroupResponse{Success: true, Users: users}, nil
}

func (r *Repository) ListUsersByGroup(ctx context.Context, groupName string) ([]*pb.User, error) {
	query := `
		SELECT u.id, u.name, u.surname, u.patronymic, u.email, u.role 
		FROM users u
		JOIN users_in_groups ug ON u.id = ug.user_id
		JOIN student_groups sg ON ug.group_id = sg.id
		WHERE sg.name = $1`
	rows, err := r.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var id int32
		var name, surname, patronymic, email, role string
		if err := rows.Scan(&id, &name, &surname, &patronymic, &email, &role); err != nil {
			return nil, err
		}
		fio := fmt.Sprintf("%s %s %s", name, surname, patronymic)
		if patronymic == "" {
			fio = fmt.Sprintf("%s %s", name, surname)
		}
		users = append(users, &pb.User{
			Id:     id,
			Fio:    fio,
			Email:  email,
			Group:  groupName,
			Status: role,
		})
	}
	return users, nil
}

func (r *Repository) ManageGroup(ctx context.Context, groupID int, action string, userID int, role string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем, существует ли группа
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM student_groups WHERE id = $1)", groupID).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("group with ID %d not found", groupID)
	}

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

	// Обновляем роль, если указано
	if role != "" {
		_, err = tx.ExecContext(ctx, "UPDATE users SET role = $1 WHERE id = $2", role, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) ListGroups(ctx context.Context) ([]*pb.Group, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT g.id, g.name, g.description FROM student_groups g")
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
		// Получаем дисциплины для группы
		var disciplines []string
		discRows, err := r.db.QueryContext(ctx, "SELECT d.name FROM disciplines d JOIN groups_in_disciplines gid ON d.id = gid.discipline_id WHERE gid.group_id = $1", id)
		if err == nil {
			defer discRows.Close()
			for discRows.Next() {
				var discName string
				if err := discRows.Scan(&discName); err == nil {
					disciplines = append(disciplines, discName)
				}
			}
		}
		groups = append(groups, &pb.Group{Id: id, Name: name, Description: description, Disciplines: disciplines})
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

func (s *Service) ListDisciplines(ctx context.Context, req *pb.ListDisciplinesRequest) (*pb.ListDisciplinesResponse, error) {
	disciplines, err := s.repo.ListDisciplines(ctx)
	if err != nil {
		return &pb.ListDisciplinesResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ListDisciplinesResponse{Success: true, Disciplines: disciplines}, nil
}

func (r *Repository) ManageDiscipline(ctx context.Context, disciplineID, groupID, seminaristID, assistantID int) error {
	query := `
		UPDATE groups_in_disciplines 
		SET seminarist_id = $1, assistant_id = $2 
		WHERE discipline_id = $3 AND group_id = $4`
	_, err := r.db.ExecContext(ctx, query, seminaristID, assistantID, disciplineID, groupID)
	return err
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

func (r *Repository) ManageDisciplineEntity(ctx context.Context, action string, groupID int32, disciplineIDs []int32, name string, seminaristID, assistantID int32) error {
	log.Printf("ManageDisciplineEntity вызван с action=%s, groupID=%d, name=%s, seminaristID=%d, assistantID=%d", action, groupID, name, seminaristID, assistantID)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if action == "attach" {
		for _, disciplineID := range disciplineIDs {
			var exists bool
			err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM groups_in_disciplines WHERE group_id = $1 AND discipline_id = $2)", groupID, disciplineID).Scan(&exists)
			if err != nil {
				return err
			}
			if !exists {
				_, err = tx.ExecContext(ctx, "INSERT INTO groups_in_disciplines (group_id, discipline_id, seminarist_id, assistant_id) VALUES ($1, $2, $3, $4)",
					groupID, disciplineID, seminaristID, assistantID)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return fmt.Errorf("недопустимое действие, должно быть 'create' или 'attach'")
	}

	return tx.Commit()
}

func (s *Service) ManageDisciplineEntity(ctx context.Context, req *pb.ManageDisciplineEntityRequest) (*pb.ManageDisciplineEntityResponse, error) {
	err := s.repo.ManageDisciplineEntity(ctx, req.Action, req.GroupId, req.DisciplineIds, req.Name, req.SeminaristId, req.AssistantId)
	if err != nil {
		return &pb.ManageDisciplineEntityResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.ManageDisciplineEntityResponse{Message: "Discipline managed successfully", Success: true}, nil
}

func (s *Service) CreateDiscipline(ctx context.Context, req *pb.ManageDisciplineEntityRequest) (*pb.ManageDisciplineEntityResponse, error) {
	log.Printf("CreateDiscipline вызван с action=%s, name=%s, groupID=%d, seminaristID=%d, assistantID=%d", req.Action, req.Name, req.GroupId, req.SeminaristId, req.AssistantId)
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return &pb.ManageDisciplineEntityResponse{Message: "Ошибка начала транзакции", Success: false}, err
	}
	defer func() {
		if err != nil {
			log.Printf("Откат транзакции: %v", tx.Rollback())
		}
	}()

	if req.Action != "create" {
		return &pb.ManageDisciplineEntityResponse{Message: "Недопустимое действие, должно быть 'create'", Success: false}, nil
	}

	if req.Name == "" {
		return &pb.ManageDisciplineEntityResponse{Message: "Для создания дисциплины требуется название", Success: false}, nil
	}

	query := "INSERT INTO disciplines (name) VALUES ($1) RETURNING id"
	var newDisciplineID int32
	err = tx.QueryRowContext(ctx, query, req.Name).Scan(&newDisciplineID)
	if err != nil {
		return &pb.ManageDisciplineEntityResponse{Message: "Ошибка вставки в disciplines", Success: false}, err
	}

	err = tx.Commit()
	if err != nil {
		return &pb.ManageDisciplineEntityResponse{Message: "Ошибка коммита транзакции", Success: false}, err
	}

	log.Printf("Дисциплина успешно создана с ID: %d", newDisciplineID)
	return &pb.ManageDisciplineEntityResponse{Message: "Дисциплина успешно создана", Success: true, DisciplineId: newDisciplineID}, nil
}

func (r *Repository) DeleteDiscipline(ctx context.Context, disciplineIDs []int32) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range disciplineIDs {
		query := "DELETE FROM disciplines WHERE id = $1"
		result, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return fmt.Errorf("discipline with ID %d not found", id)
		}
	}

	return tx.Commit()
}

func (s *Service) DeleteDiscipline(ctx context.Context, req *pb.DeleteDisciplineRequest) (*pb.DeleteDisciplineResponse, error) {
	if len(req.DisciplineIds) == 0 {
		return &pb.DeleteDisciplineResponse{Message: "at least one discipline ID is required", Success: false}, nil
	}
	err := s.repo.DeleteDiscipline(ctx, req.DisciplineIds)
	if err != nil {
		return &pb.DeleteDisciplineResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.DeleteDisciplineResponse{Message: "Disciplines deleted successfully", Success: true}, nil
}

func (r *Repository) ListDisciplines(ctx context.Context) ([]*pb.Discipline, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM disciplines")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disciplines []*pb.Discipline
	for rows.Next() {
		var id int32
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		disciplines = append(disciplines, &pb.Discipline{Id: id, Name: name})
	}
	return disciplines, nil
}

func (r *Repository) DetachDisciplinesFromGroup(ctx context.Context, groupID int32, disciplineIDs []int32) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, disciplineID := range disciplineIDs {
		query := "DELETE FROM groups_in_disciplines WHERE group_id = $1 AND discipline_id = $2"
		result, err := tx.ExecContext(ctx, query, groupID, disciplineID)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			log.Printf("No discipline with ID %d found for group %d", disciplineID, groupID)
		}
	}

	return tx.Commit()
}

// В структуре Service
func (s *Service) DetachDisciplinesFromGroup(ctx context.Context, req *pb.DetachDisciplinesFromGroupRequest) (*pb.DetachDisciplinesFromGroupResponse, error) {
	if req.GroupId <= 0 {
		return &pb.DetachDisciplinesFromGroupResponse{Message: "invalid group ID", Success: false}, nil
	}
	if len(req.DisciplineIds) == 0 {
		return &pb.DetachDisciplinesFromGroupResponse{Message: "at least one discipline ID is required", Success: false}, nil
	}

	err := s.repo.DetachDisciplinesFromGroup(ctx, req.GroupId, req.DisciplineIds)
	if err != nil {
		return &pb.DetachDisciplinesFromGroupResponse{Message: err.Error(), Success: false}, err
	}
	return &pb.DetachDisciplinesFromGroupResponse{Message: "Disciplines detached successfully", Success: true}, nil
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
