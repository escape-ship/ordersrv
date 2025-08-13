package kafka

import (
	"context"

	"github.com/escape-ship/ordersrv/pkg/postgres"
)

type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

type MessageHandler func(ctx context.Context, key, value []byte, db postgres.DBEngine) error

type Consumer interface {
	Consume(ctx context.Context)
	Close() error
}
