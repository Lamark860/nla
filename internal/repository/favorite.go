package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	SecID     string    `json:"secid"`
	CreatedAt time.Time `json:"created_at"`
}

type FavoriteRepository struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{pool: pool}
}

func (r *FavoriteRepository) Add(ctx context.Context, userID int64, secid string) error {
	query := `INSERT INTO favorites (user_id, secid) VALUES ($1, $2) ON CONFLICT (user_id, secid) DO NOTHING`
	_, err := r.pool.Exec(ctx, query, userID, secid)
	return err
}

func (r *FavoriteRepository) Remove(ctx context.Context, userID int64, secid string) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND secid = $2`
	_, err := r.pool.Exec(ctx, query, userID, secid)
	return err
}

func (r *FavoriteRepository) ListByUser(ctx context.Context, userID int64) ([]Favorite, error) {
	query := `SELECT id, user_id, secid, created_at FROM favorites WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favs []Favorite
	for rows.Next() {
		var f Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.SecID, &f.CreatedAt); err != nil {
			return nil, err
		}
		favs = append(favs, f)
	}
	return favs, rows.Err()
}

func (r *FavoriteRepository) GetSecIDs(ctx context.Context, userID int64) ([]string, error) {
	query := `SELECT secid FROM favorites WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secids []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		secids = append(secids, s)
	}
	return secids, rows.Err()
}

func (r *FavoriteRepository) IsFavorite(ctx context.Context, userID int64, secid string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND secid = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, secid).Scan(&exists)
	return exists, err
}

func (r *FavoriteRepository) CheckMultiple(ctx context.Context, userID int64, secids []string) (map[string]bool, error) {
	if len(secids) == 0 {
		return map[string]bool{}, nil
	}
	query := `SELECT secid FROM favorites WHERE user_id = $1 AND secid = ANY($2)`
	rows, err := r.pool.Query(ctx, query, userID, secids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool, len(secids))
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		result[s] = true
	}
	return result, rows.Err()
}

func (r *FavoriteRepository) Count(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}
