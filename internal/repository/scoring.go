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
	"nla/internal/scoring"
)

// ScoringRepo bundles persistence for the Phase 2 scoring engine across
// three tables: scoring_profiles, bond_scores, bond_score_explanations.
// One struct, one pool — the surface is small enough that splitting it
// would just shuffle methods around without clarifying ownership.
type ScoringRepo struct {
	pool *pgxpool.Pool
}

func NewScoringRepo(pool *pgxpool.Pool) *ScoringRepo {
	return &ScoringRepo{pool: pool}
}

// ---------- scoring_profiles ----------

// GetProfile returns one profile by code, or pgx.ErrNoRows wrapped.
func (r *ScoringRepo) GetProfile(ctx context.Context, code string) (*model.ScoringProfile, error) {
	var p model.ScoringProfile
	var weightsRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT code, name, is_preset, user_id, weights, created_at, updated_at
		FROM scoring_profiles WHERE code = $1`, code).
		Scan(&p.Code, &p.Name, &p.IsPreset, &p.UserID, &weightsRaw, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(weightsRaw, &p.Weights); err != nil {
		return nil, fmt.Errorf("decode weights for %q: %w", code, err)
	}
	return &p, nil
}

// ListProfiles returns all presets first, then user-owned profiles for the
// given userID (use 0 / nil to list only presets). Sort: presets in canonical
// low → mid → high order, then user profiles alphabetically.
func (r *ScoringRepo) ListProfiles(ctx context.Context, userID *int64) ([]model.ScoringProfile, error) {
	// CASE WHEN orders the 3 presets in canonical order without hard-coding
	// numeric weights into SQL — anything else falls back to name.
	rows, err := r.pool.Query(ctx, `
		SELECT code, name, is_preset, user_id, weights, created_at, updated_at
		FROM scoring_profiles
		WHERE is_preset = TRUE OR ($1::BIGINT IS NOT NULL AND user_id = $1)
		ORDER BY
		    is_preset DESC,
		    CASE code WHEN 'low' THEN 1 WHEN 'mid' THEN 2 WHEN 'high' THEN 3 ELSE 9 END,
		    name`, userID)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var out []model.ScoringProfile
	for rows.Next() {
		var p model.ScoringProfile
		var weightsRaw []byte
		if err := rows.Scan(&p.Code, &p.Name, &p.IsPreset, &p.UserID,
			&weightsRaw, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(weightsRaw, &p.Weights); err != nil {
			return nil, fmt.Errorf("decode weights for %q: %w", p.Code, err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---------- bond_scores ----------

// BondScore mirrors a row of the `bond_scores` table together with the
// decoded ScoreResult blob from the JSONB `breakdown` column. Lives in the
// repository package (not model) so model stays free of the scoring engine
// import, avoiding a scoring → model → scoring cycle.
type BondScore struct {
	ID          int64
	SECID       string
	ProfileCode string
	ComputedAt  time.Time
	Result      scoring.ScoreResult
}

// GetScoreByID fetches a single score row by primary key. Used by the
// worker after the explain job is enqueued — the job payload carries the
// score_id captured at enqueue time, so the worker doesn't race against
// later cache refreshes (an Insert during job latency would otherwise
// silently switch the breakdown the LLM narrates).
func (r *ScoringRepo) GetScoreByID(ctx context.Context, id int64) (*BondScore, error) {
	var s BondScore
	var breakdownRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, secid, profile_code, breakdown, computed_at
		FROM bond_scores WHERE id = $1`, id).
		Scan(&s.ID, &s.SECID, &s.ProfileCode, &breakdownRaw, &s.ComputedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get score by id: %w", err)
	}
	if err := json.Unmarshal(breakdownRaw, &s.Result); err != nil {
		return nil, fmt.Errorf("decode breakdown for id=%d: %w", id, err)
	}
	return &s, nil
}

// GetLatestScore returns the most recent cached score for (secid, profile)
// regardless of age. Service layer decides whether it's still fresh enough.
// Returns (nil, nil) when no row exists.
func (r *ScoringRepo) GetLatestScore(ctx context.Context, secid, profileCode string) (*BondScore, error) {
	var s BondScore
	var breakdownRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, secid, profile_code, breakdown, computed_at
		FROM bond_scores
		WHERE secid = $1 AND profile_code = $2
		ORDER BY computed_at DESC
		LIMIT 1`, secid, profileCode).
		Scan(&s.ID, &s.SECID, &s.ProfileCode, &breakdownRaw, &s.ComputedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest score: %w", err)
	}
	if err := json.Unmarshal(breakdownRaw, &s.Result); err != nil {
		return nil, fmt.Errorf("decode breakdown for %s/%s: %w", secid, profileCode, err)
	}
	return &s, nil
}

// InsertScore writes one freshly-computed BondScore. The full ScoreResult is
// stored in the breakdown JSONB; the top-level `score` column duplicates
// Result.Score only so it can be indexed/sorted/filtered without unpacking
// the blob.
func (r *ScoringRepo) InsertScore(ctx context.Context, s *BondScore) error {
	breakdownRaw, err := json.Marshal(s.Result)
	if err != nil {
		return fmt.Errorf("encode breakdown: %w", err)
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO bond_scores (secid, profile_code, score, breakdown)
		VALUES ($1, $2, $3, $4)
		RETURNING id, computed_at`,
		s.SECID, s.ProfileCode, s.Result.Score, breakdownRaw).
		Scan(&s.ID, &s.ComputedAt)
	if err != nil {
		return fmt.Errorf("insert score: %w", err)
	}
	return nil
}

// DeleteScoresOlderThan trims the score history to a rolling window. Not
// wired anywhere yet, but kept here so the maintenance job has a place to
// land later — the breakdown JSONB grows ~3KB per row and we don't need
// indefinite history.
func (r *ScoringRepo) DeleteScoresOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bond_scores WHERE computed_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("trim scores: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ---------- bond_score_explanations ----------

// GetExplanationByScoreID returns the cached LLM explanation for one
// BondScore.ID, or (nil, nil) if none has been generated yet.
func (r *ScoringRepo) GetExplanationByScoreID(ctx context.Context, scoreID int64) (*model.BondScoreExplanation, error) {
	var e model.BondScoreExplanation
	err := r.pool.QueryRow(ctx, `
		SELECT id, bond_score_id, llm_model, text, created_at
		FROM bond_score_explanations
		WHERE bond_score_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, scoreID).
		Scan(&e.ID, &e.BondScoreID, &e.LLMModel, &e.Text, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get explanation: %w", err)
	}
	return &e, nil
}

// InsertExplanation persists a fresh LLM-generated explanation. Multiple
// explanations for the same score are kept; the GET returns the newest.
func (r *ScoringRepo) InsertExplanation(ctx context.Context, e *model.BondScoreExplanation) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO bond_score_explanations (bond_score_id, llm_model, text)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		e.BondScoreID, e.LLMModel, e.Text).
		Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert explanation: %w", err)
	}
	return nil
}

