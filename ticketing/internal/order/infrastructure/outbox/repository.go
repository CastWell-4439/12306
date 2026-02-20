package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Event struct {
	ID          int64
	EventID     string
	AggregateID string
	EventType   string
	Payload     map[string]any
	Status      string
	CreatedAt   time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertTx(ctx context.Context, tx *sql.Tx, eventID string, aggregateID string, eventType string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO outbox(event_id, aggregate_id, event_type, payload, status) VALUES(?, ?, ?, ?, 'PENDING')`,
		eventID, aggregateID, eventType, raw,
	)
	return err
}

func (r *Repository) ListPending(ctx context.Context, limit int) ([]Event, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, event_id, aggregate_id, event_type, payload, status, created_at
		 FROM outbox
		 WHERE status='PENDING'
		 ORDER BY id ASC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Event, 0)
	for rows.Next() {
		var (
			ev  Event
			raw []byte
		)
		if err := rows.Scan(&ev.ID, &ev.EventID, &ev.AggregateID, &ev.EventType, &raw, &ev.Status, &ev.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &ev.Payload); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (r *Repository) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE outbox SET status='PUBLISHED', published_at=CURRENT_TIMESTAMP WHERE id=?`,
		id,
	)
	return err
}
