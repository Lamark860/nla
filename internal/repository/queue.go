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

// QueueRepo persists background-job state. Atomic claim via SKIP LOCKED.
type QueueRepo struct {
	pool *pgxpool.Pool
}

func NewQueueRepo(pool *pgxpool.Pool) *QueueRepo {
	return &QueueRepo{pool: pool}
}

func (r *QueueRepo) Create(ctx context.Context, job *model.QueueJob) error {
	data, err := marshalAny(job.Data)
	if err != nil {
		return fmt.Errorf("encode job data: %w", err)
	}
	result, err := marshalAny(job.Result)
	if err != nil {
		return fmt.Errorf("encode job result: %w", err)
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO queue_jobs (type, secid, reference_id, status, data, result, error, attempts, max_attempts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		job.Type, nullIfEmpty(job.SECID), nullIfEmpty(job.ReferenceID),
		job.Status, data, result, job.Error, job.Attempts, job.MaxAttempts).
		Scan(&job.JobID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (r *QueueRepo) GetByID(ctx context.Context, id string) (*model.QueueJob, error) {
	return r.scanOne(ctx, `
		SELECT id, type, COALESCE(secid,''), COALESCE(reference_id,''), status,
		       data, result, error, attempts, max_attempts,
		       created_at, updated_at, started_at, finished_at
		FROM queue_jobs WHERE id = $1`, id)
}

func (r *QueueRepo) MarkRunning(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE queue_jobs
		SET status = 'running', started_at = COALESCE(started_at, NOW())
		WHERE id = $1`, id)
	return err
}

func (r *QueueRepo) MarkDone(ctx context.Context, id string, result interface{}) error {
	raw, err := marshalAny(result)
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE queue_jobs
		SET status = 'done', result = $2, finished_at = NOW()
		WHERE id = $1`, id, raw)
	return err
}

func (r *QueueRepo) MarkError(ctx context.Context, id string, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE queue_jobs
		SET status = 'error', error = $2, finished_at = NOW()
		WHERE id = $1`, id, errMsg)
	return err
}

// FindPending returns a pending/running job with the same type+secid (dedup).
func (r *QueueRepo) FindPending(ctx context.Context, jobType string, secid string) (*model.QueueJob, error) {
	job, err := r.scanOne(ctx, `
		SELECT id, type, COALESCE(secid,''), COALESCE(reference_id,''), status,
		       data, result, error, attempts, max_attempts,
		       created_at, updated_at, started_at, finished_at
		FROM queue_jobs
		WHERE type = $1 AND secid = $2 AND status IN ('pending', 'running')
		LIMIT 1`, jobType, secid)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return job, err
}

// FetchPending atomically claims the oldest pending job and marks it running.
func (r *QueueRepo) FetchPending(ctx context.Context) (*model.QueueJob, error) {
	job, err := r.scanOne(ctx, `
		UPDATE queue_jobs
		SET status = 'running', started_at = COALESCE(started_at, NOW())
		WHERE id = (
			SELECT id FROM queue_jobs
			WHERE status = 'pending'
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, type, COALESCE(secid,''), COALESCE(reference_id,''), status,
		          data, result, error, attempts, max_attempts,
		          created_at, updated_at, started_at, finished_at`)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return job, err
}

func (r *QueueRepo) ResetStaleJobs(ctx context.Context, staleAfter time.Duration) (int64, error) {
	cutoff := time.Now().Add(-staleAfter)
	tag, err := r.pool.Exec(ctx, `
		UPDATE queue_jobs SET status = 'pending'
		WHERE status = 'running' AND updated_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("reset stale jobs: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *QueueRepo) GetStats(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM queue_jobs GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

func (r *QueueRepo) scanOne(ctx context.Context, sql string, args ...any) (*model.QueueJob, error) {
	var job model.QueueJob
	var data, result []byte
	err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&job.JobID, &job.Type, &job.SECID, &job.ReferenceID, &job.Status,
		&data, &result, &job.Error, &job.Attempts, &job.MaxAttempts,
		&job.CreatedAt, &job.UpdatedAt, &job.StartedAt, &job.FinishedAt)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &job.Data)
	}
	if len(result) > 0 {
		_ = json.Unmarshal(result, &job.Result)
	}
	return &job, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
