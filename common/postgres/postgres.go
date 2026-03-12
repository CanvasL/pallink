package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns               = int32(10)
	defaultMinConns               = int32(2)
	defaultMaxConnLifetimeSeconds = int64(1800)
	defaultMaxConnIdleSeconds     = int64(300)
	defaultHealthCheckSeconds     = int64(60)
)

// NewPool creates a pgx connection pool with sane defaults.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	maxConns := cfg.MaxConns
	if maxConns <= 0 {
		maxConns = defaultMaxConns
	}
	minConns := cfg.MinConns
	if minConns < 0 {
		minConns = 0
	}
	if minConns == 0 {
		minConns = defaultMinConns
	}

	maxLifetime := cfg.MaxConnLifetimeSeconds
	if maxLifetime <= 0 {
		maxLifetime = defaultMaxConnLifetimeSeconds
	}
	maxIdle := cfg.MaxConnIdleSeconds
	if maxIdle <= 0 {
		maxIdle = defaultMaxConnIdleSeconds
	}
	healthCheck := cfg.HealthCheckPeriodSeconds
	if healthCheck <= 0 {
		healthCheck = defaultHealthCheckSeconds
	}

	poolCfg.MaxConns = maxConns
	poolCfg.MinConns = minConns
	poolCfg.MaxConnLifetime = time.Duration(maxLifetime) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(maxIdle) * time.Second
	poolCfg.HealthCheckPeriod = time.Duration(healthCheck) * time.Second

	return pgxpool.NewWithConfig(ctx, poolCfg)
}
