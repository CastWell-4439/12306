package domain

import (
	"errors"
	"time"
)

var (
	ErrOrderViewNotFound = errors.New("order view not found")
)

type OrderView struct {
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	AmountCents   int64     `json:"amount_cents"`
	ProviderTxnID string    `json:"provider_txn_id,omitempty"`
	SeatNo        string    `json:"seat_no,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

