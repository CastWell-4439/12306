package event

import (
	"context"
	"encoding/json"
	"time"

	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/infra/order/outbox"
)

type Publisher struct {
	producer *commonkafka.Producer
	topic    string
}

func NewPublisher(producer *commonkafka.Producer, topic string) *Publisher {
	return &Publisher{
		producer: producer,
		topic:    topic,
	}
}

func (p *Publisher) Publish(ctx context.Context, ev outbox.Event) error {
	payload := map[string]any{
		"event_id":     ev.EventID,
		"aggregate_id": ev.AggregateID,
		"event_type":   ev.EventType,
		"occurred_at":  ev.CreatedAt.UTC().Format(time.RFC3339Nano),
		"payload":      ev.Payload,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, p.topic, []byte(ev.AggregateID), raw)
}


