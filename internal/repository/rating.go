package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

// RatingRepo persists issuer credit ratings. Primary key is (emitter_id, agency).
type RatingRepo struct {
	pool *pgxpool.Pool
}

func NewRatingRepo(pool *pgxpool.Pool) *RatingRepo {
	return &RatingRepo{pool: pool}
}

func (r *RatingRepo) Upsert(ctx context.Context, rating *model.IssuerRating) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO issuer_ratings (emitter_id, agency, issuer, rating, score, score_ord)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (emitter_id, agency) DO UPDATE
		SET issuer    = EXCLUDED.issuer,
		    rating    = EXCLUDED.rating,
		    score     = EXCLUDED.score,
		    score_ord = EXCLUDED.score_ord`,
		rating.EmitterID, rating.Agency, rating.Issuer,
		rating.Rating, rating.Score, rating.ScoreOrd)
	if err != nil {
		return fmt.Errorf("upsert rating: %w", err)
	}
	return nil
}

func (r *RatingRepo) GetByIssuer(ctx context.Context, issuer string) ([]model.IssuerRating, error) {
	return r.queryRatings(ctx, `
		SELECT emitter_id, agency, issuer, rating, score, score_ord, updated_at
		FROM issuer_ratings WHERE issuer = $1`, issuer)
}

func (r *RatingRepo) GetByEmitterID(ctx context.Context, emitterID int64) ([]model.IssuerRating, error) {
	return r.queryRatings(ctx, `
		SELECT emitter_id, agency, issuer, rating, score, score_ord, updated_at
		FROM issuer_ratings WHERE emitter_id = $1`, emitterID)
}

func (r *RatingRepo) GetAll(ctx context.Context) ([]model.IssuerRating, error) {
	return r.queryRatings(ctx, `
		SELECT emitter_id, agency, issuer, rating, score, score_ord, updated_at
		FROM issuer_ratings ORDER BY issuer`)
}

func (r *RatingRepo) queryRatings(ctx context.Context, sql string, args ...any) ([]model.IssuerRating, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("find ratings: %w", err)
	}
	defer rows.Close()

	var results []model.IssuerRating
	for rows.Next() {
		var rt model.IssuerRating
		if err := rows.Scan(&rt.EmitterID, &rt.Agency, &rt.Issuer,
			&rt.Rating, &rt.Score, &rt.ScoreOrd, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, rt)
	}
	return results, rows.Err()
}

func (r *RatingRepo) BulkUpsert(ctx context.Context, ratings []model.IssuerRating) error {
	if len(ratings) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, rt := range ratings {
		batch.Queue(`
			INSERT INTO issuer_ratings (emitter_id, agency, issuer, rating, score, score_ord)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (emitter_id, agency) DO UPDATE
			SET issuer    = EXCLUDED.issuer,
			    rating    = EXCLUDED.rating,
			    score     = EXCLUDED.score,
			    score_ord = EXCLUDED.score_ord`,
			rt.EmitterID, rt.Agency, rt.Issuer, rt.Rating, rt.Score, rt.ScoreOrd)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range ratings {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("bulk upsert ratings: %w", err)
		}
	}
	return nil
}

func (r *RatingRepo) Delete(ctx context.Context, issuer, agency string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM issuer_ratings WHERE issuer = $1 AND agency = $2`,
		issuer, agency)
	if err != nil {
		return fmt.Errorf("delete rating: %w", err)
	}
	return nil
}

func (r *RatingRepo) GetDistinctEmitterIDs(ctx context.Context) (map[int64]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT emitter_id FROM issuer_ratings`)
	if err != nil {
		return nil, fmt.Errorf("distinct emitter_ids: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, rows.Err()
}
