package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusInit      Status = "INIT"
	StatusReserved  Status = "RESERVED"
	StatusPaid      Status = "PAID"
	StatusTicketed  Status = "TICKETED"
	StatusCancelled Status = "CANCELLED"
)

var (
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidStateTransfer = errors.New("invalid state transition")
	ErrOrderNotFound        = errors.New("order not found")
)

type Order struct {
	OrderID        string
	IdempotencyKey string
	Status         Status
	AmountCents    int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewOrder(orderID string, idempotencyKey string, amountCents int64) (*Order, error) {
	if amountCents <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{
		OrderID:        orderID,
		IdempotencyKey: idempotencyKey,
		Status:         StatusInit,
		AmountCents:    amountCents,
	}, nil
}

func (o *Order) Reserve() error {
	if o.Status != StatusInit {
		return ErrInvalidStateTransfer
	}
	o.Status = StatusReserved
	return nil
}

func (o *Order) MarkPaid() error {
	if o.Status != StatusReserved {
		return ErrInvalidStateTransfer
	}
	o.Status = StatusPaid
	return nil
}

func (o *Order) MarkTicketed() error {
	if o.Status != StatusPaid {
		return ErrInvalidStateTransfer
	}
	o.Status = StatusTicketed
	return nil
}

func (o *Order) Cancel() error {
	if o.Status != StatusInit && o.Status != StatusReserved {
		return ErrInvalidStateTransfer
	}
	o.Status = StatusCancelled
	return nil
}
