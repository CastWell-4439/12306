package grpc_client

import (
	"context"
	"fmt"
	"time"
)

type SeatAllocatorClient interface {
	AllocateSeat(ctx context.Context, orderID string) (string, error)
}

type MockSeatAllocator struct{}

func NewMockSeatAllocator() *MockSeatAllocator {
	return &MockSeatAllocator{}
}

func (m *MockSeatAllocator) AllocateSeat(_ context.Context, orderID string) (string, error) {
	key := orderID
	if len(key) > 8 {
		key = key[:8]
	}
	return fmt.Sprintf("CARRIAGE-1-%s", key), nil
}

func DefaultTimeout() time.Duration {
	return 2 * time.Second
}

