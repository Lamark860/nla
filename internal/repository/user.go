package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"nla/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, password, name) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, user.Email, user.Password, user.Name).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, password, name, created_at, updated_at FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, password, name, created_at, updated_at FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}
