package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

const dohodCacheTTLDays = 30

// DetailsRepo persists dohod.ru parser results. The payload is large (~80 fields)
// and is stored as JSONB; only ISIN/SECID/issuer_name are indexed.
type DetailsRepo struct {
	pool *pgxpool.Pool
}

func NewDetailsRepo(pool *pgxpool.Pool) *DetailsRepo {
	return &DetailsRepo{pool: pool}
}

// Get returns cached dohod data by ISIN. Returns nil if not found or stale (>30 days).
func (r *DetailsRepo) Get(ctx context.Context, isin string) (*model.DohodBondData, error) {
	return r.fetch(ctx, `SELECT data, fetched_at, updated_at FROM dohod_details WHERE isin = $1`, isin)
}

// GetBySecid returns cached dohod data by SECID (same staleness rule).
func (r *DetailsRepo) GetBySecid(ctx context.Context, secid string) (*model.DohodBondData, error) {
	return r.fetch(ctx, `SELECT data, fetched_at, updated_at FROM dohod_details WHERE secid = $1`, secid)
}

func (r *DetailsRepo) fetch(ctx context.Context, sql string, args ...any) (*model.DohodBondData, error) {
	var raw []byte
	var fetchedAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, sql, args...).Scan(&raw, &fetchedAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find dohod details: %w", err)
	}
	if time.Since(updatedAt) > time.Duration(dohodCacheTTLDays)*24*time.Hour {
		return nil, nil
	}
	var d model.DohodBondData
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("decode dohod data: %w", err)
	}
	d.FetchedAt = fetchedAt
	d.UpdatedAt = updatedAt
	return &d, nil
}

func (r *DetailsRepo) Upsert(ctx context.Context, d *model.DohodBondData) error {
	raw, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("encode dohod data: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO dohod_details (isin, secid, issuer_name, data, fetched_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (isin) DO UPDATE
		SET secid        = EXCLUDED.secid,
		    issuer_name  = EXCLUDED.issuer_name,
		    data         = EXCLUDED.data,
		    updated_at   = NOW()`,
		d.ISIN, d.Secid, d.IssuerName, raw)
	if err != nil {
		return fmt.Errorf("upsert dohod details: %w", err)
	}
	return nil
}
