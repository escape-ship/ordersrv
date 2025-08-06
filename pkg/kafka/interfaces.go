package kafka

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

type MessageHandler func(ctx context.Context, key, value []byte) error

type Consumer interface {
	Consume(ctx context.Context)
	Close() error
}
