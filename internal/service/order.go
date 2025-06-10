package service

import (
	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	pb "github.com/escape-ship/ordersrv/proto/gen"
)

type Server struct {
	pb.OrderServiceServer
	Queries *postgresql.Queries
}

func New(query *postgresql.Queries) *Server {
	return &Server{
		Queries: query,
	}
}
