package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"ticketing-gozero/pkg/core/ticket/domain"
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

func (r *Repository) IsOrderPaidTx(ctx context.Context, tx *sql.Tx, orderID string) (bool, error) {
	row := tx.QueryRowContext(ctx, `SELECT status FROM orders WHERE order_id=?`, orderID)
	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "PAID", nil
}

func (r *Repository) InsertTicketTx(ctx context.Context, tx *sql.Tx, ticket domain.Ticket) (bool, error) {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO tickets(ticket_id, order_id, passenger_name) VALUES(?, ?, ?)`,
		ticket.TicketID, ticket.OrderID, ticket.PassengerName,
	)
	if err == nil {
		return true, nil
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return false, nil
	}
	return false, err
}

func (r *Repository) MarkOrderTicketedTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE orders SET status='TICKETED', updated_at=CURRENT_TIMESTAMP WHERE order_id=? AND status='PAID'`,
		orderID,
	)
	return err
}

