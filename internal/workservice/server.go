package workservice

import (
	"database/sql"
	Pb "rubr/proto/work"
)

type Server struct {
	Pb.UnimplementedWorkServiceServer
	Db *sql.DB
}
