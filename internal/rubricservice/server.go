package rubricservice

import (
	"database/sql"
	Pb "rubr/proto/rubric"
)

type Server struct {
	Pb.UnimplementedRubricServiceServer
	Db *sql.DB
}
