package service

import (
	"context"
	"fmt"
	"time"

	"nla/internal/model"
	"nla/internal/mongo"
)

type QueueService struct {
	repo *mongo.QueueRepo
}

func NewQueueService(repo *mongo.QueueRepo) *QueueService {
	return &QueueService{repo: repo}
}

// Enqueue creates a new job, deduplicating by type+secid if one is already pending/running
func (s *QueueService) Enqueue(ctx context.Context, jobType, secid string, data any) (*model.QueueJob, error) {
	// Dedup: check for existing pending/running job
	existing, err := s.repo.FindPending(ctx, jobType, secid)
	if err != nil {
		return nil, fmt.Errorf("find pending: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	job := &model.QueueJob{
		JobID:       generateID("job_"),
		Type:        jobType,
		SECID:       secid,
		Status:      model.JobStatusPending,
		Data:        data,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetStatus returns job status info
func (s *QueueService) GetStatus(ctx context.Context, jobID string) (*model.JobStatusResponse, error) {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return &model.JobStatusResponse{
		JobID:      job.JobID,
		Type:       job.Type,
		Status:     job.Status,
		Result:     job.Result,
		Error:      job.Error,
		CreatedAt:  job.CreatedAt,
		FinishedAt: job.FinishedAt,
	}, nil
}

// MarkDone marks job as done with result
func (s *QueueService) MarkDone(ctx context.Context, jobID string, result any) error {
	return s.repo.MarkDone(ctx, jobID, result)
}

// MarkError marks job as failed
func (s *QueueService) MarkError(ctx context.Context, jobID string, errMsg string) error {
	return s.repo.MarkError(ctx, jobID, errMsg)
}

// FetchPending atomically claims the next pending job
func (s *QueueService) FetchPending(ctx context.Context) (*model.QueueJob, error) {
	return s.repo.FetchPending(ctx)
}

// GetStats returns queue statistics
func (s *QueueService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.repo.GetStats(ctx)
}
