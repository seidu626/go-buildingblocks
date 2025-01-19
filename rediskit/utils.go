package rediskit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// Set sets a value in Redis cache
func (rk *RedisKitClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rk.cache.Set(ctx, key, value, expiration)
}

// Get retrieves a value from Redis cache
func (rk *RedisKitClient) Get(ctx context.Context, key string, dest interface{}) error {
	return HandleError(rk.cache.Get(ctx, key, dest))
}

// Delete removes a key from Redis
func (rk *RedisKitClient) Delete(ctx context.Context, key string) error {
	return HandleError(rk.client.Del(ctx, key).Err())
}

// FlushDB flushes the current database
func (rk *RedisKitClient) FlushDB(ctx context.Context) error {
	return HandleError(rk.client.FlushDB(ctx).Err())
}

// PipelineExample demonstrates pipelined commands
func (rk *RedisKitClient) PipelineExample(ctx context.Context) error {
	pipe := rk.client.Pipeline()
	incr := pipe.Incr(ctx, "counter")
	pipe.Expire(ctx, "counter", time.Hour)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	_ = incr.Val()
	return nil
}

// BulkSet sets multiple keys in a single pipeline
func (rk *RedisKitClient) BulkSet(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	pipe := rk.client.Pipeline()
	for key, value := range items {
		pipe.Set(ctx, key, value, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// BulkGet retrieves multiple keys in a single pipeline
func (rk *RedisKitClient) BulkGet(ctx context.Context, keys []string, dest map[string]interface{}) error {
	pipe := rk.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return err
	}
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			dest[keys[i]] = val
		} else if err != redis.Nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes to Redis channels and returns the PubSub
func (rk *RedisKitClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return rk.client.Subscribe(ctx, channels...)
}

// Publish publishes a message to a Redis channel
func (rk *RedisKitClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return rk.client.Publish(ctx, channel, message).Err()
}
