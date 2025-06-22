package notificationservice

import (
	"database/sql"
	Pb "rubr/proto/notification"
)

type Server struct {
	Pb.UnimplementedNotificationServiceServer
	Db *sql.DB
}
