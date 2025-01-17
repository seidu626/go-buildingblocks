package cache

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"time"

	"context"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr string, password string, db int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10000,
	})
	return &Cache{client}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false
	} else if err != nil {
		// Handle error
		return nil, false
	}
	return val, true
}

func (c *Cache) Set(key string, value []byte, expiration time.Duration) {
	ctx := context.Background()
	c.client.Set(ctx, key, value, expiration)
}
