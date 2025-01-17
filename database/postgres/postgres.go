package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seidu626/go-buildingblocks/database/dbx"
	"go.uber.org/zap"
)

// txCtx key.
type txCtx struct{}

// connCtx key.
type connCtx struct{}

// Store encapsulates the Store pool and logger.
type Store struct {
	Logger *zap.Logger
	Pool   *pgxpool.Pool
}

// NewPGXStore initializes and returns a new Store.
func NewPGXStore(logger *zap.Logger, pool *pgxpool.Pool) *Store {
	return &Store{Logger: logger, Pool: pool}
}

// TransactionContext returns a copy of the parent context which begins a transaction
// to Store.
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
func (db *Store) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		db.Logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return dbx.Errors(err)
	}
	db.Logger.Debug("Query executed successfully", zap.String("query", query))
	return nil
}

// QueryRow prepares a query expected to return a single row.
func (db *Store) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, query, args...)
}

// Query executes a SQL query and returns rows.
func (db *Store) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		db.Logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, dbx.Errors(err)
	}
	return rows, nil
}

// Close terminates the Store pool.
func (db *Store) Close() {
	db.Pool.Close()
	db.Logger.Info("PostgresDB pool closed")
}
