package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// NewPGXPool is a PostgreSQL connection pool for pgx.
//
// Usage:
// pgPool := database.NewPGXPool(context.Background(), "", &PGXStdLogger{}, tracelog.LogLevelInfo)
// defer pgPool.Close() // Close any remaining connections before shutting down your application.
//
// Instead of passing a configuration explictly with a connString,
// you might use PG environment variables such as the following to configure the database:
// PGDATABASE, PGHOST, PGPORT, PGUSER, PGPASSWORD, PGCONNECT_TIMEOUT, etc.
// Reference: https://www.postgresql.org/docs/current/libpq-envars.html
func NewPGXPool(ctx context.Context, connString string, logger tracelog.Logger, logLevel tracelog.LogLevel) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(connString) // Using environment variables instead of a connection string.
	if err != nil {
		return nil, err
	}

	conf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger,
		LogLevel: logLevel,
	}

	conf.MaxConns = 20                      // Set the maximum number of connections
	conf.MinConns = 5                       // Set the minimum number of idle connections
	conf.MaxConnLifetime = 30 * time.Minute // Recycle connections

	// pgxpool default max number of connections is the number of CPUs on your machine returned by runtime.NumCPU().
	// This number is very conservative, and you might be able to improve performance for highly concurrent applications
	// by increasing it.
	// conf.MaxConns = runtime.NumCPU() * 5
	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("pgx connection error: %w", err)
	}
	return pool, nil
}

// LogLevelFromEnv returns the tracelog.LogLevel from the environment variable PGX_LOG_LEVEL.
// By default, this is info (tracelog.LogLevelInfo), which is good for development.
// For deployments, something like tracelog.LogLevelWarn is better choice.
func LogLevelFromEnv() (tracelog.LogLevel, error) {
	if level := os.Getenv("PGX_LOG_LEVEL"); level != "" {
		l, err := tracelog.LogLevelFromString(level)
		if err != nil {
			return tracelog.LogLevelDebug, fmt.Errorf("pgx configuration: %w", err)
		}
		return l, nil
	}
	return tracelog.LogLevelInfo, nil
}

// PgErrors returns a multi-line error printing more information from *pgconn.PgError to make debugging faster.
func PgErrors(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	return fmt.Errorf(`%w
Code: %v
Detail: %v
Hint: %v
Position: %v
InternalPosition: %v
InternalQuery: %v
Where: %v
SchemaName: %v
TableName: %v
ColumnName: %v
DataTypeName: %v
ConstraintName: %v
File: %v:%v
Routine: %v`,
		err,
		pgErr.Code,
		pgErr.Detail,
		pgErr.Hint,
		pgErr.Position,
		pgErr.InternalPosition,
		pgErr.InternalQuery,
		pgErr.Where,
		pgErr.SchemaName,
		pgErr.TableName,
		pgErr.ColumnName,
		pgErr.DataTypeName,
		pgErr.ConstraintName,
		pgErr.File, pgErr.Line,
		pgErr.Routine)
}
