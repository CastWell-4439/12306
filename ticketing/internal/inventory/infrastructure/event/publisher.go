package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/inventory/infrastructure/partition"
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

func (p *Publisher) PublishMutation(ctx context.Context, record partition.MutationRecord) error {
	payload := map[string]any{
		"event_id":     uuid.NewString(),
		"aggregate_id": record.PartitionKey,
		"event_type":   string(record.EventType),
		"occurred_at":  record.OccurredAt.UTC().Format(time.RFC3339Nano),
		"payload":      record.Payload,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, p.topic, []byte(record.PartitionKey), raw)
}
