package kafka

import (
	"context"
)

type Engine interface {
	Producer() Producer
	Consumer() Consumer
	Close() error
}

type Producer interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context) (key, value []byte, err error)
	Close() error
}
