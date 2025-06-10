package main

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/escape-ship/ordersrv/internal/app"
	"github.com/escape-ship/ordersrv/internal/infra/sqlc/postgresql"
	"github.com/escape-ship/ordersrv/internal/service"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Listener 생성
	lis, err := net.Listen("tcp", ":9093")
	if err != nil {
		fmt.Println("failed to listen:", err)
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

	queries := postgresql.New(db)
	orderController := service.NewOrderController(db, queries)

	brokers := []string{"localhost:9092"}
	topic := "order-events"
	groupID := "order-group"
	engine := kafka.NewEngine(brokers, topic, groupID)
	consumer := engine.Consumer()

	// App 인스턴스 생성 및 실행
	application := app.NewApp(db, lis, orderController, engine, consumer)
	application.Run()
}
