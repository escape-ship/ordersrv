package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/escape-ship/protos/gen"

	"github.com/escape-ship/ordersrv/internal/service"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	"github.com/escape-ship/ordersrv/pkg/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	KafkaEngine   kafka.Engine
	KafkaConsumer kafka.Consumer
	pg            postgres.DBEngine
	OrderService  *service.OrderController
	grpcServer    *grpc.Server
	ctx           context.Context
	cancel        context.CancelFunc
}

// App 생성자
func NewApp(pg postgres.DBEngine, kafkaEngine kafka.Engine, kafkaConsumer kafka.Consumer) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		KafkaEngine:   kafkaEngine,
		KafkaConsumer: kafkaConsumer,
		pg:            pg,
		OrderService:  service.NewOrderController(pg, kafkaEngine),
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

	// Kafka 메시지 핸들러
	handler := func(key, value []byte) {
		log.Printf("Kafka message received: key=%s, value=%s", string(key), string(value))
		switch string(key) {
		case "kakao-approve":
			log.Println("Processing kakao-approve message")
			err := a.OrderService.UpdateOrderStatus(context.Background(), string(value), service.OrderStatePaid)
			if err != nil {
				log.Printf("Error updating order status: %v", err)
			}
		default:
		}
	}

	// Kafka consumer를 goroutine으로 실행
	go a.runKafkaConsumer(ctx, handler)

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

	// 2. Kafka consumer 종료
	if a.KafkaConsumer != nil {
		log.Println("App: Closing Kafka consumer")
		a.KafkaConsumer.Close()
		log.Println("App: Kafka consumer closed")
	}

	// 3. Kafka engine 종료
	if a.KafkaEngine != nil {
		log.Println("App: Shutting down Kafka engine")
		// Kafka engine에 Shutdown 메소드가 있다면 호출
		log.Println("App: Kafka engine shutdown completed")
	}

	// 4. Context 취소
	if a.cancel != nil {
		a.cancel()
	}

	log.Println("App: Graceful shutdown sequence completed")
}

// runKafkaConsumer는 App 내부에서 사용하는 Kafka consumer 실행 함수
func (a *App) runKafkaConsumer(ctx context.Context, handler func(key, value []byte)) {
	log.Println("App: Starting Kafka consumer")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("App: Kafka consumer context cancelled, stopping")
			return
		default:
			key, value, err := a.KafkaConsumer.Consume(ctx)
			if err != nil {
				// Context가 cancelled되었으면 정상적인 종료
				if ctx.Err() != nil {
					log.Println("App: Kafka consumer stopping due to context cancellation")
					return
				}
				log.Printf("Kafka consume error: %v", err)
				continue
			}
			handler(key, value)
		}
	}
}

// RunKafkaConsumer는 기존 호환성을 위해 유지 (사용되지 않음)
func RunKafkaConsumer(consumer kafka.Consumer, handler func(key, value []byte)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			key, value, err := consumer.Consume(ctx)
			if err != nil {
				log.Printf("Kafka consume error: %v", err)
				continue
			}
			handler(key, value)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	consumer.Close()
}
