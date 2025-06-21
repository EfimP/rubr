package workassignmentservice

import (
	"database/sql"
	Pb "rubr/proto/workassignment"
)

type Server struct {
	Pb.UnimplementedWorkAssignmentServiceServer
	Db *sql.DB
}
