package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"ticketing-gozero/pkg/core/order/domain"
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

func (r *Repository) FindByID(ctx context.Context, orderID string) (*domain.Order, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT order_id, idempotency_key, status, amount_cents, created_at, updated_at
		 FROM orders WHERE order_id=?`,
		orderID,
	)
	return scanOrder(row)
}

func (r *Repository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT order_id, idempotency_key, status, amount_cents, created_at, updated_at
		 FROM orders WHERE idempotency_key=?`,
		key,
	)
	return scanOrder(row)
}

func (r *Repository) InsertTx(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO orders(order_id, idempotency_key, status, amount_cents) VALUES(?, ?, ?, ?)`,
		order.OrderID, order.IdempotencyKey, string(order.Status), order.AmountCents,
	)
	if err == nil {
		return nil
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return fmt.Errorf("duplicate: %w", err)
	}
	return err
}

func (r *Repository) UpdateStatusTx(ctx context.Context, tx *sql.Tx, orderID string, expected domain.Status, next domain.Status) (bool, error) {
	res, err := tx.ExecContext(
		ctx,
		`UPDATE orders SET status=?, updated_at=CURRENT_TIMESTAMP WHERE order_id=? AND status=?`,
		string(next), orderID, string(expected),
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *Repository) InsertPaymentTx(ctx context.Context, tx *sql.Tx, paymentID string, orderID string, providerTxnID string, status string) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO payments(payment_id, order_id, provider_txn_id, status) VALUES(?, ?, ?, ?)`,
		paymentID, orderID, providerTxnID, status,
	)
	return err
}

func scanOrder(row interface {
	Scan(dest ...any) error
}) (*domain.Order, error) {
	o := &domain.Order{}
	var status string
	if err := row.Scan(&o.OrderID, &o.IdempotencyKey, &status, &o.AmountCents, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	o.Status = domain.Status(status)
	return o, nil
}


