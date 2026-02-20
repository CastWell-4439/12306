package ttl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

const (
	delayQueueKey = "inventory:holds:delay_queue"
	holdKeyPrefix = "inventory:hold:"
)

type HoldValue struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
	Qty          int    `json:"qty"`
}

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

func (s *Store) Save(ctx context.Context, value HoldValue) error {
	key := holdKey(value.HoldID)
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	expAt := time.Now().Add(s.ttl).Unix()
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, key, raw, s.ttl)
	pipe.ZAdd(ctx, delayQueueKey, redisv9.Z{
		Score:  float64(expAt),
		Member: value.HoldID,
	})
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) Remove(ctx context.Context, holdID string) error {
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, holdKey(holdID))
	pipe.ZRem(ctx, delayQueueKey, holdID)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) PollExpired(ctx context.Context, limit int64) ([]HoldValue, error) {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	ids, err := s.redis.ZRangeByScore(ctx, delayQueueKey, &redisv9.ZRangeBy{
		Min:   "-inf",
		Max:   now,
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	values := make([]HoldValue, 0, len(ids))
	for _, holdID := range ids {
		raw, err := s.redis.Get(ctx, holdKey(holdID)).Bytes()
		if err == redisv9.Nil {
			_, _ = s.redis.ZRem(ctx, delayQueueKey, holdID).Result()
			continue
		}
		if err != nil {
			return nil, err
		}
		var v HoldValue
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func holdKey(holdID string) string {
	return fmt.Sprintf("%s%s", holdKeyPrefix, holdID)
}
