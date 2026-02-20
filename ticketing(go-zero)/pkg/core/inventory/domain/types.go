package domain

import "errors"

type EventType string

const (
	EventTypeHoldCreated   EventType = "hold_created"
	EventTypeHoldReleased  EventType = "hold_released"
	EventTypeHoldConfirmed EventType = "hold_confirmed"
)

var (
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrHoldNotFound      = errors.New("hold not found")
	ErrBackpressure      = errors.New("wal backpressure")
)

type Hold struct {
	HoldID string `json:"hold_id"`
	Qty    int    `json:"qty"`
}

type PartitionState struct {
	PartitionKey string          `json:"partition_key"`
	Capacity     int             `json:"capacity"`
	Available    int             `json:"available"`
	Confirmed    int             `json:"confirmed"`
	LastSeq      int64           `json:"last_seq"`
	Holds        map[string]Hold `json:"holds"`
}

func NewPartitionState(partitionKey string, capacity int) *PartitionState {
	return &PartitionState{
		PartitionKey: partitionKey,
		Capacity:     capacity,
		Available:    capacity,
		Confirmed:    0,
		LastSeq:      0,
		Holds:        map[string]Hold{},
	}
}
