package rediskit

import (
	"context"
)

// Ping checks the Redis connection
func (rk *RedisKitClient) Ping(ctx context.Context) error {
	return rk.client.Ping(ctx).Err()
}
