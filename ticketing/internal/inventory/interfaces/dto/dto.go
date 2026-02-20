package dto

type TryHoldRequest struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
	Qty          int    `json:"qty"`
	Capacity     int    `json:"capacity"`
}

type ReleaseHoldRequest struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
}

type ConfirmHoldRequest struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
}
