package domain

import "errors"

var (
	ErrOrderNotPaid = errors.New("order is not in PAID status")
)

type Ticket struct {
	TicketID      string
	OrderID       string
	PassengerName string
}


