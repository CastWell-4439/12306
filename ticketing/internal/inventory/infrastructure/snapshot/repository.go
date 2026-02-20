package snapshot

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"ticketing/internal/inventory/domain"
)

type Record struct {
	PartitionKey string
	SnapshotSeq  int64
	State        *domain.PartitionState
	CreatedAt    time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(ctx context.Context, rec Record) error {
	stateBlob, err := json.Marshal(rec.State)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO inventory_snapshot(partition_key, snapshot_seq, state_blob)
		 VALUES(?, ?, ?)
		 ON DUPLICATE KEY UPDATE snapshot_seq=VALUES(snapshot_seq), state_blob=VALUES(state_blob), created_at=CURRENT_TIMESTAMP`,
		rec.PartitionKey, rec.SnapshotSeq, stateBlob,
	)
	return err
}

func (r *Repository) LoadAll(ctx context.Context) ([]Record, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT partition_key, snapshot_seq, state_blob, created_at
		 FROM inventory_snapshot`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Record, 0)
	for rows.Next() {
		var (
			key       string
			seq       int64
			stateRaw  []byte
			createdAt time.Time
		)
		if err := rows.Scan(&key, &seq, &stateRaw, &createdAt); err != nil {
			return nil, err
		}
		st := &domain.PartitionState{}
		if err := json.Unmarshal(stateRaw, st); err != nil {
			return nil, err
		}
		out = append(out, Record{
			PartitionKey: key,
			SnapshotSeq:  seq,
			State:        st,
			CreatedAt:    createdAt,
		})
	}
	return out, rows.Err()
}
