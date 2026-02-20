package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(addr string, password string, dbIndex int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func HealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

