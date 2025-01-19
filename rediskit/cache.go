package rediskit

import (
	"context"
	"github.com/go-redis/cache/v9"
	"time"
)

// RedisCache wraps the go-redis/cache with a selected encoder
type RedisCache struct {
	cache   *cache.Cache
	encoder Encoder
}

// NewRedisCache initializes a new RedisCache with the provided Redis client and encoder
func NewRedisCache(client UniversalClient, encoder Encoder, defaultExpiration time.Duration) *RedisCache {
	return &RedisCache{
		cache: cache.New(&cache.Options{
			Redis:      client,
			Marshal:    encoder.Marshal,
			Unmarshal:  encoder.Unmarshal,
			LocalCache: cache.NewTinyLFU(1000, defaultExpiration), // optional local cache
		}),
		encoder: encoder,
	}
}

// Set sets a value in the cache with the specified key and expiration
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   expiration,
	})
}

// Get retrieves a value from the cache and unmarshals it into dest
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	return HandleError(rc.cache.Get(ctx, key, dest))
}

// Delete removes a key from the cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.cache.Delete(ctx, key)
}
