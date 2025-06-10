package app

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/escape-ship/ordersrv/internal/service"
	"github.com/escape-ship/ordersrv/pkg/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	GRPCServer      *grpc.Server
	OrderController *service.OrderController
	KafkaEngine     kafka.Engine
	KafkaConsumer   kafka.Consumer
	DB              *sql.DB
	Listener        net.Listener
}

// App 생성자
func NewApp(db *sql.DB, listener net.Listener, orderController *service.OrderController, kafkaEngine kafka.Engine, kafkaConsumer kafka.Consumer) *App {
	grpcServer := grpc.NewServer()
	// gRPC 서비스 등록
	// pb.RegisterOrderServiceServer(grpcServer, orderController) // <- main.go에서 이미 등록했다면 주석처리
	reflection.Register(grpcServer)
	return &App{
		GRPCServer:      grpcServer,
		OrderController: orderController,
		KafkaEngine:     kafkaEngine,
		KafkaConsumer:   kafkaConsumer,
		DB:              db,
		Listener:        listener,
	}
}

// App 실행: gRPC 서버와 Kafka consumer를 모두 실행
func (a *App) Run() {
	// Kafka 메시지 핸들러
	handler := func(key, value []byte) {
		log.Printf("Kafka message received: key=%s, value=%s", string(key), string(value))
		// TODO: 메시지에 따라 비즈니스 로직 실행
		switch string(key) {
		case "kakao-approve":
			// 인벤토리 감소 함수 호출
			log.Println("Processing kakao-approve message")
		default:
		}
	}
	go RunKafkaConsumer(a.KafkaConsumer, handler)

	log.Println("gRPC server listening on :9093")
	if err := a.GRPCServer.Serve(a.Listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Kafka consumer를 실행하고 메시지 처리 핸들러를 등록하는 함수
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
