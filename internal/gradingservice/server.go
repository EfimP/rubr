package gradingservice

import (
	"database/sql"
	Pb "rubr/proto/grade"
)

type Server struct {
	Pb.UnimplementedGradingServiceServer
	Db *sql.DB
}
