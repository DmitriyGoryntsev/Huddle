package postgres

import (
	"auth-service/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type DB struct {
	Pool   *pgxpool.Pool
	logger *zap.SugaredLogger
	cb     *gobreaker.CircuitBreaker
}

func NewPostgres(cfg *config.PostgresConfig, logger *zap.SugaredLogger) (*DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	logger.Infof("Connecting to Postgres with Host: %s, Port: %s, DB: %s", cfg.Host, cfg.Port, cfg.DBName)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
		if err != nil {
			logger.Warnf("Failed to create pool on attempt %d: %v", attempt, err)
			if attempt == cfg.MaxRetries {
				return nil, fmt.Errorf("failed to create Postgres pool after %d attempts: %w", cfg.MaxRetries, err)
			}
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err == nil {
			logger.Infof("Connected to Postgres on attempt %d", attempt)
			cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
				Name:        "postgres",
				MaxRequests: 1,
				Interval:    30 * time.Second,
				Timeout:     10 * time.Second,
				ReadyToTrip: func(counts gobreaker.Counts) bool {
					return counts.ConsecutiveFailures >= 3
				},
				OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
					logger.Infof("%s circuit breaker state changed: %s -> %s", name, from.String(), to.String())
				},
			})
			return &DB{Pool: pool, logger: logger, cb: cb}, nil
		}

		logger.Warnf("Postgres connection failed on attempt %d, retrying in %ds", attempt, cfg.RetryDelay)
		pool.Close()
		if attempt < cfg.MaxRetries {
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect to Postgres after %d attempts", cfg.MaxRetries)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) Ping(ctx context.Context) error {
	_, err := db.cb.Execute(func() (interface{}, error) {
		return nil, db.Pool.Ping(ctx)
	})
	return err
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	result, err := db.cb.Execute(func() (interface{}, error) {
		return db.Pool.QueryRow(ctx, sql, args...), nil
	})

	if err != nil {
		db.logger.Errorf("Circuit Breaker rejected QueryRow: %v", err)
		return &errorRow{err: err}
	}

	return result.(pgx.Row)
}

func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := db.cb.Execute(func() (interface{}, error) {
		_, err := db.Pool.Exec(ctx, sql, args...)
		return nil, err
	})

	return err
}

type errorRow struct {
	err error
}

func (r *errorRow) Scan(dest ...interface{}) error {
	return r.err
}
