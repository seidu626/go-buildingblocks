package dbx

import (
	"context"
	"errors"
	"fmt"
	"github.com/seidu626/go-buildingblocks/database"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// NewDBXPool is a PostgreSQL connection pool for pgx.
//
// Usage:
// pgPool := NewDBXPool(context.Background(), *zap.Logger, &DBXStdLogger{}, tracelog.LogLevelInfo, database.Config)
// defer pgPool.Close() // Close any remaining connections before shutting down your application.
//
// Instead of passing a configuration explicitly with a connString,
// you might use PG environment variables such as the following to configure the database:
// PGDATABASE, PGHOST, PGPORT, PGUSER, PGPASSWORD, PGCONNECT_TIMEOUT, etc.
// Reference: https://www.postgresql.org/docs/current/libpq-envars.html
func NewDBXPool(ctx context.Context, logger *zap.Logger, traceLogger tracelog.Logger, logLevel tracelog.LogLevel, config *database.Config) (*pgxpool.Pool, error) {
	dsn := buildDSN(config)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("Failed to parse CockroachDB DSN", zap.Error(err))
		return nil, err
	}

	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   traceLogger,
		LogLevel: logLevel,
	}

	// pgxpool default max number of connections is the number of CPUs on your machine returned by runtime.NumCPU().
	// This number is very conservative, and you might be able to improve performance for highly concurrent applications
	// by increasing it.
	// conf.MaxConns = runtime.NumCPU() * 5
	poolConfig.MaxConns = int32(config.PoolSize.Max)
	poolConfig.MinConns = int32(config.PoolSize.Min)
	poolConfig.MaxConnIdleTime = config.ConnMaxLifetime
	poolConfig.MaxConnLifetime = config.ConnMaxLifetime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("pgx connection error: %w", err)
	}

	return pool, nil
}

// buildDSN constructs the Data Source Name based on the config.
func buildDSN(config *database.Config) string {
	hosts := strings.Join(config.Hosts, ",")
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		hosts,
		config.Port,
		config.Database,
	)

	if config.SSL {
		dsn += "?sslmode=verify-full"
		// Append SSL parameters if needed
	} else {
		dsn += "?sslmode=disable"
	}

	return dsn
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

// StdLogger prints pgx logs to the standard logger.
// os.Stderr by default.
type StdLogger struct{}

func (l *StdLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	args := make([]any, 0, len(data)+2) // making space for arguments + level + msg
	args = append(args, level, msg)
	for k, v := range data {
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	log.Println(args...)
}

// Errors returns a multi-line error printing more information from *pgconn.PgError to make debugging faster.
func Errors(err error) error {
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
