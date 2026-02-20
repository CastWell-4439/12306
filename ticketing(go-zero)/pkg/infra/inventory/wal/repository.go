package wal

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"ticketing-gozero/pkg/core/inventory/domain"
	"ticketing-gozero/pkg/core/inventory/partition"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Append(ctx context.Context, rec partition.MutationRecord) error {
	payload, err := json.Marshal(rec.Payload)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO inventory_wal(partition_key, seq, event_type, payload, occurred_at) VALUES(?, ?, ?, ?, ?)`,
		rec.PartitionKey, rec.Seq, string(rec.EventType), payload, rec.OccurredAt,
	)
	return err
}

func (r *Repository) LoadAll(ctx context.Context) ([]partition.MutationRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT partition_key, seq, event_type, payload, occurred_at
		 FROM inventory_wal
		 ORDER BY partition_key ASC, seq ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]partition.MutationRecord, 0)
	for rows.Next() {
		var (
			key        string
			seq        int64
			eventType  string
			payloadRaw []byte
			occurredAt time.Time
		)
		if err := rows.Scan(&key, &seq, &eventType, &payloadRaw, &occurredAt); err != nil {
			return nil, err
		}
		var payload map[string]any
		if err := json.Unmarshal(payloadRaw, &payload); err != nil {
			return nil, err
		}
		out = append(out, partition.MutationRecord{
			PartitionKey: key,
			Seq:          seq,
			EventType:    domain.EventType(eventType),
			Payload:      payload,
			OccurredAt:   occurredAt,
		})
	}
	return out, rows.Err()
}

