package database

import "time"

// Config holds the configuration parameters for CockroachDB connection.
type Config struct {
	Hosts           []string
	Port            int
	Username        string
	Password        string
	Database        string
	SSL             bool
	PoolSize        PoolSize
	ConnectTimeout  time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type PoolSize struct {
	Max int
	Min int
}
