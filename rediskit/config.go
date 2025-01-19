package rediskit

import (
	"crypto/tls"
	"time"
)

// Config holds the configuration for the RedisKit client
type Config struct {
	// Redis connection settings
	Addr         string        // Redis server address (e.g., "localhost:6379")
	Password     string        // Password for Redis
	DB           int           // Database number
	PoolSize     int           // Maximum number of socket connections
	MinIdleConns int           // Minimum number of idle connections
	IdleTimeout  time.Duration // Idle timeout duration

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	RetryCount   int
	RetryDelay   time.Duration

	// Retry settings
	MaxRetries      int           // Maximum number of retries
	MinRetryBackoff time.Duration // Minimum backoff between retries
	MaxRetryBackoff time.Duration // Maximum backoff between retries

	// Cache settings
	DefaultExpiration time.Duration // Default cache expiration

	// Encoding type: "json", "msgpack", "protobuf"
	Encoding string

	// Cluster settings
	IsCluster bool
	Addrs     []string // Addresses for Redis Cluster

	// TLS settings (optional)
	TLSConfig *tls.Config
}
