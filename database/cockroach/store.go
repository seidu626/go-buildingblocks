package cockroach

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seidu626/go-buildingblocks/database"
	"github.com/seidu626/go-buildingblocks/database/dbx"
	"go.uber.org/zap"
	"time"
)

// txCtx key.
type txCtx struct{}

// connCtx key.
type connCtx struct{}

// Store encapsulates the CockroachDB pool and logger.
type Store struct {
	Logger *zap.Logger
	Pool   *pgxpool.Pool
	Config *database.Config
}

// NewCockroachStore initializes and returns a new Store.
func NewCockroachStore(logger *zap.Logger, pool *pgxpool.Pool, config *database.Config) *Store {
	return &Store{Logger: logger, Pool: pool, Config: config}
}

// TransactionContext returns a copy of the parent context which begins a transaction
// to CockroachDB.
//
// Once the transaction is over, you must call db.Commit(ctx) to make the changes effective.
// This might live in the go-pkg/postgres package later for the sake of code reuse.
func (db *Store) TransactionContext(ctx context.Context) (context.Context, error) {
	tx, err := db.Conn(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, txCtx{}, tx), nil
}

// Commit transaction from context.
func (db *Store) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value(txCtx{}).(pgx.Tx); ok && tx != nil {
		return tx.Commit(ctx)
	}
	return errors.New("context has no transaction")
}

// Rollback transaction from context.
func (db *Store) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value(txCtx{}).(pgx.Tx); ok && tx != nil {
		return tx.Rollback(ctx)
	}
	return errors.New("context has no transaction")
}

// WithAcquire returns a copy of the parent context which acquires a connection
// to Store from pgxpool to make sure commands executed in series reuse the
// same database connection.
//
// To release the connection back to the pool, you must call postgres.Release(ctx).
//
// Example:
// dbCtx := db.WithAcquire(ctx)
// defer postgres.Release(dbCtx)
func (db *Store) WithAcquire(ctx context.Context) (dbCtx context.Context, err error) {
	if _, ok := ctx.Value(connCtx{}).(*pgxpool.Conn); ok {
		panic("context already has a connection acquired")
	}
	res, err := db.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, connCtx{}, res), nil
}

// Release Store connection acquired by context back to the pool.
func (db *Store) Release(ctx context.Context) {
	if res, ok := ctx.Value(connCtx{}).(*pgxpool.Conn); ok && res != nil {
		res.Release()
	}
}

// Conn returns a Store transaction if one exists.
// If not, returns a connection if a connection has been acquired by calling WithAcquire.
// Otherwise, it returns *pgxpool.Pool which acquires the connection and closes it immediately after a SQL command is executed.
func (db *Store) Conn(ctx context.Context) dbx.Querier {
	if tx, ok := ctx.Value(txCtx{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	if res, ok := ctx.Value(connCtx{}).(*pgxpool.Conn); ok && res != nil {
		return res
	}
	return db.Pool
}

// Exec executes a SQL query without returning any rows.
func (db *Store) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	cmdTag, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		db.Logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return cmdTag, dbx.Errors(err)
	}
	db.Logger.Debug("Query executed successfully", zap.String("query", query))
	return cmdTag, nil
}

// QueryRow prepares a query expected to return a single row.
func (db *Store) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return db.Pool.QueryRow(ctx, query, args...)
}

// Query executes a SQL query and returns rows.
func (db *Store) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		db.Logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, dbx.Errors(err)
	}
	return rows, nil
}

func (db *Store) MonitorAndAdjustPool() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Example: Monitor active connections
			stats := db.Pool.Stat()
			activeConns := stats.AcquiredConns()

			if float64(activeConns) > float64(db.Config.MaxOpenConns)*0.8 {
				// Increase pool size
				db.Config.MaxOpenConns += 5
				db.Pool.Config().MaxConns = int32(db.Config.MaxOpenConns)
				db.Logger.Info("Increased pool size", zap.Int("new_pool_size", db.Config.MaxOpenConns))
			} else if float64(activeConns) < float64(db.Config.MaxOpenConns)*0.2 && db.Config.MaxOpenConns > 10 {
				// Decrease pool size
				db.Config.MaxOpenConns -= 5
				db.Pool.Config().MaxConns = int32(db.Config.MaxOpenConns)
				db.Logger.Info("Decreased pool size", zap.Int("new_pool_size", db.Config.MaxOpenConns))
			}

			// Apply new pool configuration if necessary
			// Note: Some drivers may require pool recreation to apply new settings
		}
	}
}

// Close terminates the CockroachDB pool.
func (db *Store) Close() {
	db.Pool.Close()
	db.Logger.Info("CockroachDB pool closed")
}
