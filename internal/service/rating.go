package service

import (
	"context"
	"fmt"

	"nla/internal/model"
	mongoRepo "nla/internal/mongo"
)

type RatingService struct {
	repo *mongoRepo.RatingRepo
}

func NewRatingService(repo *mongoRepo.RatingRepo) *RatingService {
	return &RatingService{repo: repo}
}

func (s *RatingService) GetByIssuer(ctx context.Context, issuer string) (*model.IssuerRatingResponse, error) {
	ratings, err := s.repo.GetByIssuer(ctx, issuer)
	if err != nil {
		return nil, err
	}

	resp := &model.IssuerRatingResponse{
		Issuer:  issuer,
		Ratings: ratings,
	}

	if len(ratings) > 0 {
		maxScore := 0
		for _, r := range ratings {
			if r.Score > maxScore {
				maxScore = r.Score
			}
		}
		resp.Score = maxScore
		resp.EmitterID = ratings[0].EmitterID
	}

	return resp, nil
}

func (s *RatingService) GetByEmitterID(ctx context.Context, emitterID int64) (*model.IssuerRatingResponse, error) {
	ratings, err := s.repo.GetByEmitterID(ctx, emitterID)
	if err != nil {
		return nil, err
	}

	resp := &model.IssuerRatingResponse{
		EmitterID: emitterID,
		Ratings:   ratings,
	}

	if len(ratings) > 0 {
		maxScore := 0
		for _, r := range ratings {
			if r.Score > maxScore {
				maxScore = r.Score
			}
			if resp.Issuer == "" {
				resp.Issuer = r.Issuer
			}
		}
		resp.Score = maxScore
	}

	return resp, nil
}

// GetAll returns all ratings grouped by emitter_id (key = emitter_id as string)
func (s *RatingService) GetAll(ctx context.Context) (map[string]*model.IssuerRatingResponse, error) {
	ratings, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*model.IssuerRatingResponse)
	for _, r := range ratings {
		key := fmt.Sprintf("%d", r.EmitterID)
		resp, ok := result[key]
		if !ok {
			resp = &model.IssuerRatingResponse{
				EmitterID: r.EmitterID,
				Issuer:    r.Issuer,
				Ratings:   []model.IssuerRating{},
			}
			result[key] = resp
		}
		resp.Ratings = append(resp.Ratings, r)
		if r.Score > resp.Score {
			resp.Score = r.Score
		}
	}

	return result, nil
}

func (s *RatingService) Upsert(ctx context.Context, rating *model.IssuerRating) error {
	return s.repo.Upsert(ctx, rating)
}

func (s *RatingService) BulkUpsert(ctx context.Context, ratings []model.IssuerRating) error {
	return s.repo.BulkUpsert(ctx, ratings)
}

func (s *RatingService) Delete(ctx context.Context, issuer, agency string) error {
	return s.repo.Delete(ctx, issuer, agency)
}

// SeedDefaults is a no-op. Ratings are now populated dynamically from dohod.ru.
func (s *RatingService) SeedDefaults(ctx context.Context) error {
	return nil
}
