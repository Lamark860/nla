package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

// AnalysisRepo persists bond AI analyses. Each record carries the full LLM response,
// a parsed numeric rating (0-100), plus arbitrary JSON inputs.
type AnalysisRepo struct {
	pool *pgxpool.Pool
}

func NewAnalysisRepo(pool *pgxpool.Pool) *AnalysisRepo {
	return &AnalysisRepo{pool: pool}
}

func (r *AnalysisRepo) Save(ctx context.Context, a *model.BondAnalysis) error {
	jsonData, err := marshalAny(a.JSONData)
	if err != nil {
		return fmt.Errorf("encode json_data: %w", err)
	}
	customJSON, err := marshalAny(a.CustomJSON)
	if err != nil {
		return fmt.Errorf("encode custom_json: %w", err)
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO bond_analyses (secid, response, rating, json_data, custom_json, user_id, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, timestamp, saved_at`,
		a.SECID, a.Response, a.Rating, jsonData, customJSON, a.UserID, a.Tags).
		Scan(&a.ID, &a.Timestamp, &a.SavedAt)
	if err != nil {
		return fmt.Errorf("save analysis: %w", err)
	}
	return nil
}

func (r *AnalysisRepo) GetBySecid(ctx context.Context, secid string) ([]model.BondAnalysis, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, secid, response, rating, json_data, custom_json, user_id, timestamp, saved_at, tags
		FROM bond_analyses WHERE secid = $1
		ORDER BY timestamp DESC`, secid)
	if err != nil {
		return nil, fmt.Errorf("find analyses: %w", err)
	}
	defer rows.Close()

	var results []model.BondAnalysis
	for rows.Next() {
		var a model.BondAnalysis
		var jsonData, customJSON []byte
		if err := rows.Scan(&a.ID, &a.SECID, &a.Response, &a.Rating,
			&jsonData, &customJSON, &a.UserID, &a.Timestamp, &a.SavedAt, &a.Tags); err != nil {
			return nil, err
		}
		if err := unmarshalAny(jsonData, &a.JSONData); err != nil {
			return nil, err
		}
		if err := unmarshalAny(customJSON, &a.CustomJSON); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

func (r *AnalysisRepo) GetByID(ctx context.Context, id string) (*model.BondAnalysis, error) {
	var a model.BondAnalysis
	var jsonData, customJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, secid, response, rating, json_data, custom_json, user_id, timestamp, saved_at, tags
		FROM bond_analyses WHERE id = $1`, id).
		Scan(&a.ID, &a.SECID, &a.Response, &a.Rating,
			&jsonData, &customJSON, &a.UserID, &a.Timestamp, &a.SavedAt, &a.Tags)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("analysis not found")
		}
		return nil, fmt.Errorf("analysis not found: %w", err)
	}
	if err := unmarshalAny(jsonData, &a.JSONData); err != nil {
		return nil, err
	}
	if err := unmarshalAny(customJSON, &a.CustomJSON); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AnalysisRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bond_analyses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete analysis: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("analysis not found")
	}
	return nil
}

func (r *AnalysisRepo) GetStats(ctx context.Context, secid string) (*model.AnalysisStats, error) {
	stats := &model.AnalysisStats{}
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*),
		       COALESCE(AVG(rating), 0),
		       MAX(timestamp)
		FROM bond_analyses WHERE secid = $1`, secid).
		Scan(&stats.Total, &stats.AvgRating, &stats.LastAnalysis)
	if err != nil {
		return nil, fmt.Errorf("get analysis stats: %w", err)
	}
	return stats, nil
}

// GetLatestRatings returns the most recent rating per SECID for the given batch.
func (r *AnalysisRepo) GetLatestRatings(ctx context.Context, secids []string) (map[string]float64, error) {
	if len(secids) == 0 {
		return map[string]float64{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (secid) secid, rating
		FROM bond_analyses
		WHERE secid = ANY($1) AND rating IS NOT NULL
		ORDER BY secid, timestamp DESC`, secids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var secid string
		var rating float64
		if err := rows.Scan(&secid, &rating); err != nil {
			return nil, err
		}
		result[secid] = rating
	}
	return result, rows.Err()
}

// GetBulkStats returns analysis count + avg rating + last timestamp per SECID.
func (r *AnalysisRepo) GetBulkStats(ctx context.Context) (map[string]model.AnalysisStats, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT secid, COUNT(*), COALESCE(AVG(rating), 0), MAX(timestamp)
		FROM bond_analyses
		GROUP BY secid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.AnalysisStats)
	for rows.Next() {
		var secid string
		var stats model.AnalysisStats
		if err := rows.Scan(&secid, &stats.Total, &stats.AvgRating, &stats.LastAnalysis); err != nil {
			return nil, err
		}
		result[secid] = stats
	}
	return result, rows.Err()
}

func marshalAny(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

func unmarshalAny(raw []byte, dst *any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}
