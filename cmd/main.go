package main

import (
	"database/sql"
	"fmt"
	"net"

	pb "github.com/escape-ship/ordersrv/proto/gen"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":9093")
	if err != nil {
		return
	}

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?parseTime=true",
		"testuser", "testpasswd", "0.0.0.0", "5432", "escape")

	fmt.Println("Connecting to DB:", dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// m, err := migrate.New("file://db/migrations", dsn)
	// if err != nil {
	// 	log.Fatal("Migration init failed:", err)
	// }
	// if err := m.Up(); err != nil && err != migrate.ErrNoChange {
	// 	log.Fatal("Migration failed:", err)
	// }
	// fmt.Println("Database migrated successfully!")

	queries := postgresql.New(db)

	s := grpc.NewServer()

	pb.RegisterOrderServer(s, &server{queries: queries})

	reflection.Register(s)

	fmt.Println("Serving ordersrv on http://0.0.0.0:8081")

	if err := s.Serve(lis); err != nil {
		return
	}
}
