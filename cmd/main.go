package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	brokers := []string{"kafka:9092"}
	topic := "payments"
	groupID := "order-group"
	engine := kafka.NewEngine(brokers, topic, groupID)
	consumer := engine.Consumer()

	// App 인스턴스 생성
	application := app.NewApp(db, engine, consumer)

	// Context와 signal handling 설정
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 시그널 채널 설정
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// App 실행 (goroutine으로)
	go func() {
		if err := application.Run(ctx); err != nil {
			logger.Error("App: application run error", "error", err)
			cancel()
		}
	}()

	// 시그널 대기
	select {
	case sig := <-sigChan:
		logger.Info("App: received signal, starting graceful shutdown", "signal", sig)
	case <-ctx.Done():
		logger.Info("App: context cancelled, starting graceful shutdown")
	}

	// Graceful shutdown 실행
	logger.Info("App: graceful shutdown sequence started")
	application.Shutdown()
	logger.Info("App: graceful shutdown completed")
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
