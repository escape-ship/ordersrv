package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type engine struct {
	producer Producer
	consumer Consumer
}

func NewEngine(brokers []string, topic, groupID string, opts ...Option) Engine {
	e := &engine{}
	for _, opt := range opts {
		opt(e)
	}
	e.producer = newProducer(brokers, topic)
	e.consumer = newConsumer(brokers, topic, groupID)
	return e
}

func (e *engine) Configure(opts ...Option) Engine {
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *engine) Producer() Producer {
	return e.producer
}

func (e *engine) Consumer() Consumer {
	return e.consumer
}

func (e *engine) Close() error {
	if e.producer != nil {
		e.producer.Close()
	}
	if e.consumer != nil {
		e.consumer.Close()
	}
	return nil
}

// Producer implementation

type producer struct {
	writer *kafka.Writer
}

func newProducer(brokers []string, topic string) Producer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &producer{writer: w}
}

func (p *producer) Publish(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{Key: key, Value: value}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *producer) Close() error {
	return p.writer.Close()
}

// Consumer implementation

type consumer struct {
	reader *kafka.Reader
}

func newConsumer(brokers []string, topic, groupID string) Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &consumer{reader: r}
}

func (c *consumer) Consume(ctx context.Context) (key, value []byte, err error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, nil, err
	}
	return msg.Key, msg.Value, nil
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
