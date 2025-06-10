package publisher

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Publisher wraps kafka.Writer for producing messages
type Publisher struct {
	writer *kafka.Writer
}

var _ interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
} = (*Publisher)(nil)

// NewPublisher creates a new Kafka publisher
func NewPublisher(brokers []string, topic string) *Publisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &Publisher{writer: writer}
}

// Publish writes a message to Kafka
func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Close closes the publisher
func (p *Publisher) Close() error {
	return p.writer.Close()
}
