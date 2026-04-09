package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id          BIGSERIAL PRIMARY KEY,
			email       VARCHAR(255) NOT NULL UNIQUE,
			password    VARCHAR(255) NOT NULL,
			name        VARCHAR(255) NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)`,
		`CREATE TABLE IF NOT EXISTS favorites (
			id         BIGSERIAL PRIMARY KEY,
			user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			secid      VARCHAR(64) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, secid)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id)`,
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
