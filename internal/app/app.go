package app

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/escape-ship/protos/gen"

	"github.com/escape-ship/ordersrv/internal/service"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	"github.com/escape-ship/ordersrv/pkg/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	KafkaConsumer []kafka.Consumer
	pg            postgres.DBEngine
	OrderService  *service.OrderController
	grpcServer    *grpc.Server
	ctx           context.Context
	cancel        context.CancelFunc
}

// App 생성자
func NewApp(pg postgres.DBEngine, kafkaConsumer []kafka.Consumer) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		KafkaConsumer: kafkaConsumer,
		pg:            pg,
		OrderService:  service.NewOrderController(pg),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// App 실행: gRPC 서버와 Kafka consumer를 모두 실행
func (a *App) Run(ctx context.Context) error {
	// gRPC 서버 설정
	a.grpcServer = grpc.NewServer()
	pb.RegisterOrderServiceServer(a.grpcServer, a.OrderService)
	reflection.Register(a.grpcServer)

	// Listener 생성
	lis, err := net.Listen("tcp", ":8083")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Kafka consumer를 goroutine으로 실행
	for _, consumer := range a.KafkaConsumer {
		go consumer.Consume(ctx)
	}

	// gRPC 서버를 goroutine으로 실행
	go func() {
		log.Println("gRPC server listening on :8083")
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Context 완료 대기
	<-ctx.Done()
	return nil
}

// Shutdown 메소드: graceful shutdown 수행
func (a *App) Shutdown() {
	log.Println("App: Starting graceful shutdown sequence")

	// 1. gRPC 서버 graceful stop
	if a.grpcServer != nil {
		log.Println("App: Stopping gRPC server")
		a.grpcServer.GracefulStop()
		log.Println("App: gRPC server stopped")
	}

	// 4. Context 취소
	if a.cancel != nil {
		a.cancel()
	}

	log.Println("App: Graceful shutdown sequence completed")
}
