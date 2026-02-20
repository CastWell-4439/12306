package dto

type CreateOrderRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	AmountCents    int64  `json:"amount_cents"`
}

type ReserveOrderRequest struct {
	OrderID      string `json:"order_id"`
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
	Qty          int    `json:"qty"`
	Capacity     int    `json:"capacity"`
}

type PaymentCallbackRequest struct {
	OrderID       string `json:"order_id"`
	ProviderTxnID string `json:"provider_txn_id"`
	Status        string `json:"status"`
	PartitionKey  string `json:"partition_key"`
	HoldID        string `json:"hold_id"`
	Signature     string `json:"signature"`
}

type CancelOrderRequest struct {
	OrderID      string `json:"order_id"`
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
}
