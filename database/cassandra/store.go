package cassandra

import (
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

// CxStore encapsulates the Cassandra session and logger.
type CxStore struct {
	Logger   *zap.Logger
	Session  *gocql.Session
	Cluster  *gocql.ClusterConfig
	Keyspace string
}

// Config holds the configuration parameters for Cassandra connection.
type Config struct {
	Hosts          []string
	Port           int
	Username       string
	Password       string
	Keyspace       string
	Consistency    gocql.Consistency
	Timeout        time.Duration
	ConnectTimeout time.Duration
	PoolSize       int
	RetryPolicy    gocql.RetryPolicy
	ProtoVersion   int
	SSL            bool
	SslOpts        *gocql.SslOptions // Updated to use SslOptions
}

// NewCassandraStore initializes and returns a new CassandraStore.
func NewCassandraStore(logger *zap.Logger, config *Config) (*CxStore, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Port = config.Port
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = config.Consistency
	cluster.Timeout = config.Timeout
	cluster.ConnectTimeout = config.ConnectTimeout
	cluster.PoolConfig.HostSelectionPolicy = gocql.RoundRobinHostPolicy()
	cluster.NumConns = config.PoolSize
	cluster.RetryPolicy = config.RetryPolicy
	cluster.ProtoVersion = config.ProtoVersion

	if config.Username != "" && config.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	if config.SSL && config.SslOpts != nil {
		cluster.SslOpts = config.SslOpts
	}

	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error("Failed to create Cassandra session", zap.Error(err))
		return nil, err
	}

	store := &CxStore{
		Logger:   logger,
		Session:  session,
		Cluster:  cluster,
		Keyspace: config.Keyspace,
	}

	logger.Info("Cassandra session established successfully")

	return store, nil
}

// QueryRow executes a CQL query expected to return a single row.
func (c *CxStore) QueryRow(query string, values ...interface{}) *gocql.Query {
	return c.Session.Query(query, values...)
}

// IterateRows executes a CQL query and iterates over the returned rows.
func (c *CxStore) IterateRows(query string, iterFunc func(*gocql.Iter) error, values ...interface{}) error {
	iter := c.Session.Query(query, values...).Iter()
	for iter.Scan() {
		// The iter.Scan() call here doesn't scan into any variables.
		// It's up to the iterFunc to handle data extraction.
		if err := iterFunc(iter); err != nil {
			c.Logger.Error("Error processing iterator", zap.Error(err))
			return err
		}
	}
	if err := iter.Close(); err != nil {
		c.Logger.Error("Iterator closed with error", zap.Error(err))
		return err
	}
	return nil
}

// Exec executes a CQL query without returning any rows.
func (c *CxStore) Exec(query string, values ...interface{}) error {
	if err := c.Session.Query(query, values...).Exec(); err != nil {
		c.Logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return err
	}
	c.Logger.Debug("Query executed successfully", zap.String("query", query))
	return nil
}

// Close terminates the Cassandra session.
func (c *CxStore) Close() {
	c.Session.Close()
	c.Logger.Info("Cassandra session closed")
}
