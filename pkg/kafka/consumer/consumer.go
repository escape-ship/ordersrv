package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps kafka.Reader for consuming messages
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: reader}
}

// Consume reads a single message from Kafka
func (c *Consumer) Consume(ctx context.Context) (key, value []byte, err error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, nil, err
	}
	return msg.Key, msg.Value, nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
