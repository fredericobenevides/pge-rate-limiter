package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorageStrategy struct {
	Client *redis.Client
}

func NewRedisStorageStrategy(client *redis.Client) *RedisStorageStrategy {
	return &RedisStorageStrategy{
		Client: client,
	}
}

func (rs *RedisStorageStrategy) AllowRequest(ctx context.Context, id string, requestLimit int, windowDuration time.Duration, blockDuration time.Duration) (bool, error) {
	windowDurationKey := fmt.Sprintf("rate_limiter:window:%s", id)
	blockDurationKey := fmt.Sprintf("rate_limiter:block:%s", id)

	blocked, err := rs.Client.Get(ctx, blockDurationKey).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if blocked != "" {
		return false, nil
	}

	count, err := rs.Client.Incr(ctx, windowDurationKey).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		rs.Client.Expire(ctx, windowDurationKey, windowDuration)
	} else if count > int64(requestLimit) {
		rs.Client.Set(ctx, blockDurationKey, true, blockDuration)
		return false, nil
	}

	return true, nil
}
