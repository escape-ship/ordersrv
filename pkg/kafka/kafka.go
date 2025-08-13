package kafka

import (
	"context"

	"github.com/escape-ship/ordersrv/pkg/postgres"
	"github.com/segmentio/kafka-go"
)

// Publisher implementation

type publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers []string, topic string) Publisher {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &publisher{writer: w}
}

func (p *publisher) Publish(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{Key: key, Value: value}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *publisher) Close() error {
	return p.writer.Close()
}

// Consumer implementation

type consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	db      postgres.DBEngine
}

func NewConsumer(brokers []string, topics map[string]MessageHandler, groupID string, db postgres.DBEngine) []Consumer {
	var res []Consumer
	for topic, handler := range topics {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		})
		res = append(res, &consumer{reader: r, handler: handler, db: db})
	}
	return res
}

func (c *consumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				return
			}
			if err := c.handler(ctx, msg.Key, msg.Value, c.db); err != nil {
				// Handle error from message handler
				continue
			}
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
