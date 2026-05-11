package service

import (
	"context"
	"fmt"
	"log"

	"nla/internal/model"
	"nla/internal/repository"
)

type RatingService struct {
	repo *repository.RatingRepo
}

func NewRatingService(repo *repository.RatingRepo) *RatingService {
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
	fillScores(rating)
	return s.repo.Upsert(ctx, rating)
}

func (s *RatingService) BulkUpsert(ctx context.Context, ratings []model.IssuerRating) error {
	for i := range ratings {
		fillScores(&ratings[i])
	}
	return s.repo.BulkUpsert(ctx, ratings)
}

// fillScores ensures ScoreOrd / Score are derived from Rating text whenever the
// caller didn't supply them. Centralising this in the service layer keeps every
// write path (HTTP handler, dohod sync, MOEX CCI sync, recompute) in agreement.
func fillScores(r *model.IssuerRating) {
	if r.Rating == "" || r.ScoreOrd != 0 {
		return
	}
	r.ScoreOrd, _ = NormalizeRating(r.Rating)
	if r.Score == 0 {
		r.Score = LegacyScore10(r.ScoreOrd)
	}
}

// RecomputeAllScores walks every existing rating and rewrites Score/ScoreOrd
// using the current NormalizeRating implementation. Intended to be called once
// at API startup so that data persisted under the old (buggy) ratingToScore
// mapping is brought up to date.
//
// Records whose Rating text NormalizeRating cannot parse (ord == 0) are left
// untouched so that any manually-entered legacy Score is not silently zeroed.
func (s *RatingService) RecomputeAllScores(ctx context.Context) (int, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return 0, err
	}
	updated := 0
	for i := range all {
		ord, _ := NormalizeRating(all[i].Rating)
		if ord == 0 {
			continue
		}
		legacy := LegacyScore10(ord)
		if all[i].ScoreOrd == ord && all[i].Score == legacy {
			continue
		}
		all[i].ScoreOrd = ord
		all[i].Score = legacy
		if err := s.repo.Upsert(ctx, &all[i]); err != nil {
			log.Printf("[rating-recompute] WARN: emitter %d agency %q: %v", all[i].EmitterID, all[i].Agency, err)
			continue
		}
		updated++
	}
	return updated, nil
}

func (s *RatingService) Delete(ctx context.Context, issuer, agency string) error {
	return s.repo.Delete(ctx, issuer, agency)
}

// SeedDefaults is a no-op. Ratings are now populated dynamically from dohod.ru.
func (s *RatingService) SeedDefaults(ctx context.Context) error {
	return nil
}
