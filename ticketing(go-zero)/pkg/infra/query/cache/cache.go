package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisv9 "github.com/redis/go-redis/v9"

	"ticketing-gozero/pkg/core/query/domain"
)

type Store struct {
	redis *redisv9.Client
	ttl   time.Duration
}

func NewStore(redis *redisv9.Client, ttl time.Duration) *Store {
	return &Store{
		redis: redis,
		ttl:   ttl,
	}
}

func (s *Store) GetOrderView(ctx context.Context, orderID string) (*domain.OrderView, bool, error) {
	raw, err := s.redis.Get(ctx, keyOrderView(orderID)).Bytes()
	if err == redisv9.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out domain.OrderView
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func (s *Store) SetOrderView(ctx context.Context, v *domain.OrderView) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, keyOrderView(v.OrderID), raw, s.ttl).Err()
}

func (s *Store) InvalidateOrderView(ctx context.Context, orderID string) error {
	return s.redis.Del(ctx, keyOrderView(orderID)).Err()
}

func keyOrderView(orderID string) string {
	return fmt.Sprintf("query:order:%s", orderID)
}

