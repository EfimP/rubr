package userservice

import (
	"database/sql"
	Pb "rubr/proto/user"
)

type Server struct {
	Pb.UnimplementedUserServiceServer
	Db *sql.DB
}
