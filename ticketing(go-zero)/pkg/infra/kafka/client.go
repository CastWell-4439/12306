package kafka

import (
	"context"
	"net"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *segmentkafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &segmentkafka.Writer{
			Addr:         segmentkafka.TCP(brokers...),
			Balancer:     &segmentkafka.Hash{},
			RequiredAcks: segmentkafka.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte) error {
	return p.writer.WriteMessages(ctx, segmentkafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *segmentkafka.Reader
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: segmentkafka.NewReader(segmentkafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 1 * time.Second,
		}),
	}
}

func (c *Consumer) Read(ctx context.Context) (segmentkafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Fetch(ctx context.Context) (segmentkafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) Commit(ctx context.Context, msg segmentkafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func HealthCheck(ctx context.Context, brokers []string) error {
	if len(brokers) == 0 {
		return net.ErrClosed
	}
	d := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}

