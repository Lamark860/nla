package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

// IssuerRepo persists SECID → emitter mapping. Postgres analogue of the
// former internal/mongo.IssuerRepo. Method signatures are kept identical
// so service layer doesn't need to know which backing store is used.
type IssuerRepo struct {
	pool *pgxpool.Pool
}

func NewIssuerRepo(pool *pgxpool.Pool) *IssuerRepo {
	return &IssuerRepo{pool: pool}
}

func (r *IssuerRepo) GetAll(ctx context.Context) ([]model.BondIssuer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT secid, emitter_id, emitter_name, is_hidden, needs_sync, created_at, updated_at
		FROM bond_issuers`)
	if err != nil {
		return nil, fmt.Errorf("find issuers: %w", err)
	}
	defer rows.Close()

	var results []model.BondIssuer
	for rows.Next() {
		var b model.BondIssuer
		if err := rows.Scan(&b.SECID, &b.EmitterID, &b.EmitterName,
			&b.IsHidden, &b.NeedsSync, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *IssuerRepo) GetBySecid(ctx context.Context, secid string) (*model.BondIssuer, error) {
	var b model.BondIssuer
	err := r.pool.QueryRow(ctx, `
		SELECT secid, emitter_id, emitter_name, is_hidden, needs_sync, created_at, updated_at
		FROM bond_issuers WHERE secid = $1`, secid).
		Scan(&b.SECID, &b.EmitterID, &b.EmitterName,
			&b.IsHidden, &b.NeedsSync, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find issuer by secid: %w", err)
	}
	return &b, nil
}

func (r *IssuerRepo) GetBySecids(ctx context.Context, secids []string) (map[string]model.BondIssuer, error) {
	if len(secids) == 0 {
		return map[string]model.BondIssuer{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT secid, emitter_id, emitter_name, is_hidden, needs_sync, created_at, updated_at
		FROM bond_issuers WHERE secid = ANY($1)`, secids)
	if err != nil {
		return nil, fmt.Errorf("find issuers by secids: %w", err)
	}
	defer rows.Close()

	result := make(map[string]model.BondIssuer, len(secids))
	for rows.Next() {
		var b model.BondIssuer
		if err := rows.Scan(&b.SECID, &b.EmitterID, &b.EmitterName,
			&b.IsHidden, &b.NeedsSync, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		result[b.SECID] = b
	}
	return result, rows.Err()
}

func (r *IssuerRepo) ToggleHidden(ctx context.Context, emitterID int64, hidden bool) (int64, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE bond_issuers SET is_hidden = $1 WHERE emitter_id = $2`,
		hidden, emitterID)
	if err != nil {
		return 0, fmt.Errorf("toggle issuer %d: %w", emitterID, err)
	}
	return tag.RowsAffected(), nil
}

func (r *IssuerRepo) GetHiddenEmitterIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT emitter_id FROM bond_issuers WHERE is_hidden = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// UpdateEmitterName fills emitter_name for rows that still have it empty.
// Mirrors Mongo behaviour: only updates rows where emitter_name = ''.
func (r *IssuerRepo) UpdateEmitterName(ctx context.Context, emitterID int64, name string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bond_issuers SET emitter_name = $1
		WHERE emitter_id = $2 AND emitter_name = ''`,
		name, emitterID)
	if err != nil {
		return fmt.Errorf("update emitter name %d: %w", emitterID, err)
	}
	return nil
}

func (r *IssuerRepo) Upsert(ctx context.Context, issuer *model.BondIssuer) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bond_issuers (secid, emitter_id, emitter_name, is_hidden, needs_sync)
		VALUES ($1, $2, $3, FALSE, TRUE)
		ON CONFLICT (secid) DO UPDATE
		SET emitter_id   = EXCLUDED.emitter_id,
		    emitter_name = EXCLUDED.emitter_name`,
		issuer.SECID, issuer.EmitterID, issuer.EmitterName)
	if err != nil {
		return fmt.Errorf("upsert issuer %s: %w", issuer.SECID, err)
	}
	return nil
}

func (r *IssuerRepo) GetAllSecids(ctx context.Context) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT secid FROM bond_issuers`)
	if err != nil {
		return nil, fmt.Errorf("find all secids: %w", err)
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		result[s] = true
	}
	return result, rows.Err()
}

// GetOneSecidPerEmitter returns one sample secid for each non-hidden emitter.
// Used for bulk rating sync — one bond per emitter is enough to fetch dohod.ru.
func (r *IssuerRepo) GetOneSecidPerEmitter(ctx context.Context) (map[int64]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (emitter_id) emitter_id, secid
		FROM bond_issuers
		WHERE emitter_id > 0 AND is_hidden = FALSE
		ORDER BY emitter_id, secid`)
	if err != nil {
		return nil, fmt.Errorf("aggregate emitters: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var id int64
		var secid string
		if err := rows.Scan(&id, &secid); err != nil {
			return nil, err
		}
		if id > 0 && secid != "" {
			result[id] = secid
		}
	}
	return result, rows.Err()
}
