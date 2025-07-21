package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/escape-ship/ordersrv/config"
	"github.com/escape-ship/ordersrv/internal/app"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	"github.com/escape-ship/ordersrv/pkg/postgres"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx 드라이버 등록
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.New("config.yaml")
	if err != nil {
		logger.Error("App: config load error", "error", err)
		os.Exit(1)
	}

	db, err := postgres.New(makeDSN(cfg.Database))
	if err != nil {
		logger.Error("App: database connection error", "error", err)
		os.Exit(1)
	}

	brokers := []string{"localhost:9092"}
	topic := "payments"
	groupID := "order-group"
	engine := kafka.NewEngine(brokers, topic, groupID)
	consumer := engine.Consumer()

	// App 인스턴스 생성 및 실행
	application := app.NewApp(db, engine, consumer)
	application.Run()
}

// config.Database 값 사용
func makeDSN(db config.Database) postgres.DBConnString {
	return postgres.DBConnString(
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
			db.User, db.Password,
			db.Host, db.Port,
			db.DataBaseName, db.SSLMode, db.SchemaName,
		),
	)
}
