package readmodel

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"ticketing-gozero/pkg/core/query/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *sql.DB {
	return r.db
}

func (r *Repository) GetOrderView(ctx context.Context, orderID string) (*domain.OrderView, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT order_id, status, amount_cents, provider_txn_id, seat_no, updated_at
		 FROM query_order_view
		 WHERE order_id=?`,
		orderID,
	)
	var v domain.OrderView
	if err := row.Scan(&v.OrderID, &v.Status, &v.AmountCents, &v.ProviderTxnID, &v.SeatNo, &v.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderViewNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *Repository) UpsertOrderViewTx(ctx context.Context, tx *sql.Tx, v domain.OrderView) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO query_order_view(order_id, status, amount_cents, provider_txn_id, seat_no, updated_at)
		 VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON DUPLICATE KEY UPDATE
		   status=VALUES(status),
		   amount_cents=VALUES(amount_cents),
		   provider_txn_id=VALUES(provider_txn_id),
		   seat_no=VALUES(seat_no),
		   updated_at=CURRENT_TIMESTAMP`,
		v.OrderID, v.Status, v.AmountCents, v.ProviderTxnID, v.SeatNo,
	)
	return err
}

func (r *Repository) MarkConsumedTx(ctx context.Context, tx *sql.Tx, eventID string, consumer string) (bool, error) {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO consumed_events(event_id, consumer_name, consumed_at) VALUES(?, ?, CURRENT_TIMESTAMP)`,
		eventID, consumer,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) UpdateOrderFromOrderEventTx(
	ctx context.Context,
	tx *sql.Tx,
	orderID string,
	status string,
	amountCents int64,
	providerTxnID string,
) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO query_order_view(order_id, status, amount_cents, provider_txn_id, seat_no, updated_at)
		 VALUES(?, ?, ?, ?, '', CURRENT_TIMESTAMP)
		 ON DUPLICATE KEY UPDATE
		   status=VALUES(status),
		   amount_cents=CASE
		     WHEN VALUES(amount_cents) > 0 THEN VALUES(amount_cents)
		     ELSE amount_cents
		   END,
		   provider_txn_id=CASE
		     WHEN VALUES(provider_txn_id) <> '' THEN VALUES(provider_txn_id)
		     ELSE provider_txn_id
		   END,
		   updated_at=CURRENT_TIMESTAMP`,
		orderID, status, amountCents, providerTxnID,
	)
	return err
}

func (r *Repository) MarkTicketedTx(ctx context.Context, tx *sql.Tx, orderID string, seatNo string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO query_order_view(order_id, status, amount_cents, provider_txn_id, seat_no, updated_at)
		 VALUES(?, 'TICKETED', 0, '', ?, CURRENT_TIMESTAMP)
		 ON DUPLICATE KEY UPDATE
		   status='TICKETED',
		   seat_no=VALUES(seat_no),
		   updated_at=CURRENT_TIMESTAMP`,
		orderID, seatNo,
	)
	return err
}

func (r *Repository) RebuildFromOrders(ctx context.Context, limit int) error {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT order_id, status, amount_cents, updated_at FROM orders ORDER BY updated_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			orderID     string
			status      string
			amountCents int64
			updatedAt   time.Time
		)
		if err := rows.Scan(&orderID, &status, &amountCents, &updatedAt); err != nil {
			return err
		}
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO query_order_view(order_id, status, amount_cents, provider_txn_id, seat_no, updated_at)
			 VALUES(?, ?, ?, '', '', ?)
			 ON DUPLICATE KEY UPDATE
			   status=VALUES(status),
			   amount_cents=VALUES(amount_cents),
			   updated_at=VALUES(updated_at)`,
			orderID, status, amountCents, updatedAt,
		)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}


