package rediskit

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// UniversalClient is an interface that encompasses both Redis and Cluster clients
type UniversalClient interface {
	redis.Cmdable
	redis.ClusterCmdable
	Ping(ctx context.Context) *redis.StatusCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Close() error
}

// RedisKitClient wraps the Redis client and cache
type RedisKitClient struct {
	client        UniversalClient
	cache         *RedisCache
	config        Config
	isCluster     bool
	encoder       Encoder
	clusterClient *redis.ClusterClient
}

// NewRedisKitClient initializes and returns a RedisKitClient
func NewRedisKitClient(cfg Config) (*RedisKitClient, error) {
	var client UniversalClient
	var err error

	// Select encoder based on configuration
	encoder, err := SelectEncoder(cfg.Encoding)
	if err != nil {
		return nil, err
	}

	if cfg.IsCluster {
		if len(cfg.Addrs) == 0 {
			return nil, fmt.Errorf("cluster mode enabled but no addresses provided")
		}
		clusterOptions := &redis.ClusterOptions{
			Addrs:           cfg.Addrs,
			Password:        cfg.Password,
			MaxRetries:      cfg.MaxRetries,
			MinRetryBackoff: cfg.MinRetryBackoff,
			MaxRetryBackoff: cfg.MaxRetryBackoff,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			//IdleTimeout:     cfg.IdleTimeout,
			TLSConfig: cfg.TLSConfig,
		}
		clusterClient := redis.NewClusterClient(clusterOptions)
		if err := clusterClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis cluster: %v", err)
		}
		client = clusterClient
	} else {
		options := &redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			//IdleTimeout:     cfg.IdleTimeout,
			MaxRetries:      cfg.MaxRetries,
			MinRetryBackoff: cfg.MinRetryBackoff,
			MaxRetryBackoff: cfg.MaxRetryBackoff,
			TLSConfig:       cfg.TLSConfig,
		}
		redisClient := redis.NewClient(options)
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %v", err)
		}
		client = redisClient
	}

	// Initialize cache
	cacheInstance := NewRedisCache(client, encoder, cfg.DefaultExpiration)

	return &RedisKitClient{
		client:    client,
		cache:     cacheInstance,
		config:    cfg,
		isCluster: cfg.IsCluster,
		encoder:   encoder,
	}, nil
}

// Close gracefully closes the Redis client
func (rk *RedisKitClient) Close() error {
	return rk.client.Close()
}
