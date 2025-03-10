package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/escape-ship/ordersrv/internal/sql/mysql"
	pb "github.com/escape-ship/ordersrv/proto/gen"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.OrderServer
	queries *mysql.Queries
}

// Insert implements helloworld.GreeterServer
func (s *server) Insert(ctx context.Context, in *pb.InsertRequestMessage) (*pb.InsertResponseMessage, error) {
	result, err := s.queries.CreateOrder(
		ctx,
		mysql.CreateOrderParams{
			OrderSource:     in.GetOrderSource(),
			LoyaltyMemberID: in.GetLoyaltyMemberId(),
			OrderStatus:     in.GetOrderStatus(),
			Updated:         sql.NullTime{Time: time.Now(), Valid: true},
		},
	)

	if err != nil {
		return nil, err
	}

	r, _ := result.RowsAffected()
	fmt.Println("affected rows:", r)

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &pb.InsertResponseMessage{Id: id}, nil
}

// GetAll implements helloworld.GreeterServer
func (s *server) GetAll(ctx context.Context, _ *pb.GetAllRequestMessage) (*pb.GetAllResponseMessage, error) {
	orders, err := s.queries.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	orderList := []*pb.OrderMessage{}
	for _, order := range orders {
		orderList = append(orderList, &pb.OrderMessage{
			Id:              order.ID,
			OrderSource:     order.OrderSource,
			LoyaltyMemberId: order.LoyaltyMemberID,
			OrderStatus:     order.OrderStatus,
			Updated:         order.Updated.Time.String(),
		})
	}

	return &pb.GetAllResponseMessage{Orders: orderList}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return
	}

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?parseTime=true",
		"testuser", "testpassword", "0.0.0.0", "3306", "escape")

	fmt.Println("Connecting to DB:", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatal("Migration init failed:", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("Database migrated successfully!")

	queries := mysql.New(db)

	s := grpc.NewServer()

	pb.RegisterOrderServer(s, &server{queries: queries})

	reflection.Register(s)

	fmt.Println("Serving ordersrv on http://0.0.0.0:8081")

	if err := s.Serve(lis); err != nil {
		return
	}
}
