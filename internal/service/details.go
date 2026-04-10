package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"nla/internal/client/dohod"
	"nla/internal/model"
	mongorepo "nla/internal/mongo"
)

const (
	dohodMaxRetries   = 3
	dohodBaseDelay    = 1 * time.Second
	JobTypeParseDohod = "parse_dohod"
)

type DetailsService struct {
	dohodClient *dohod.Client
	detailsRepo *mongorepo.DetailsRepo
	ratingRepo  *mongorepo.RatingRepo
	issuerRepo  *mongorepo.IssuerRepo
}

func NewDetailsService(dohodClient *dohod.Client, detailsRepo *mongorepo.DetailsRepo, ratingRepo *mongorepo.RatingRepo, issuerRepo *mongorepo.IssuerRepo) *DetailsService {
	return &DetailsService{
		dohodClient: dohodClient,
		detailsRepo: detailsRepo,
		ratingRepo:  ratingRepo,
		issuerRepo:  issuerRepo,
	}
}

// GetDetails returns dohod.ru details for a bond. Returns cached if fresh.
func (s *DetailsService) GetDetails(ctx context.Context, secid, isin string) (*model.DohodBondData, error) {
	// Try cache first (by ISIN, then by secid)
	if isin != "" {
		if cached, err := s.detailsRepo.Get(ctx, isin); err == nil && cached != nil {
			// Ensure ratings are synced from cached data
			s.updateRatingsFromDohod(ctx, cached)
			return cached, nil
		}
	}
	if secid != "" {
		if cached, err := s.detailsRepo.GetBySecid(ctx, secid); err == nil && cached != nil {
			// Ensure ratings are synced from cached data
			s.updateRatingsFromDohod(ctx, cached)
			return cached, nil
		}
	}

	if isin == "" {
		return nil, fmt.Errorf("ISIN required to fetch dohod.ru data")
	}

	return nil, nil // no cache — needs async fetch
}

// FetchAndSave fetches data from dohod.ru with retry and saves to MongoDB.
// Called by queue worker.
func (s *DetailsService) FetchAndSave(ctx context.Context, secid, isin string) (*model.DohodBondData, error) {
	var lastErr error

	for attempt := 0; attempt < dohodMaxRetries; attempt++ {
		if attempt > 0 {
			delay := dohodBaseDelay * time.Duration(1<<uint(attempt-1)) // 1s, 2s, 4s
			log.Printf("Dohod retry %d/%d for %s after %v", attempt+1, dohodMaxRetries, isin, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		data, err := s.dohodClient.FetchBond(ctx, isin)
		if err != nil {
			lastErr = err
			log.Printf("Dohod fetch error for %s (attempt %d): %v", isin, attempt+1, err)
			continue
		}

		data.Secid = secid

		if err := s.detailsRepo.Upsert(ctx, data); err != nil {
			return nil, fmt.Errorf("save dohod data: %w", err)
		}

		// Update issuer ratings from dohod.ru data
		s.updateRatingsFromDohod(ctx, data)

		return data, nil
	}

	return nil, fmt.Errorf("dohod fetch failed after %d attempts: %w", dohodMaxRetries, lastErr)
}

// updateRatingsFromDohod extracts credit ratings from dohod.ru data and updates issuer_ratings collection.
// Uses bond_issuers to resolve emitter_id by secid.
func (s *DetailsService) updateRatingsFromDohod(ctx context.Context, data *model.DohodBondData) {
	if s.ratingRepo == nil || s.issuerRepo == nil {
		return
	}

	// Resolve emitter_id from bond_issuers
	issuerInfo, err := s.issuerRepo.GetBySecid(ctx, data.Secid)
	if err != nil || issuerInfo == nil || issuerInfo.EmitterID == 0 {
		log.Printf("WARN: cannot resolve emitter_id for secid %s, skipping rating update", data.Secid)
		return
	}

	emitterID := issuerInfo.EmitterID
	issuerName := data.IssuerName
	if issuerName == "" {
		issuerName = issuerInfo.EmitterName
	}

	// Update emitter_name in bond_issuers if it was empty and we now have a name
	if issuerInfo.EmitterName == "" && issuerName != "" {
		if err := s.issuerRepo.UpdateEmitterName(ctx, emitterID, issuerName); err != nil {
			log.Printf("WARN: update emitter name for %d: %v", emitterID, err)
		}
	}

	var ratings []model.IssuerRating
	if data.AKRA != "" {
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "АКРА",
			Rating:    data.AKRA,
			Score:     ratingToScore(data.AKRA),
		})
	}
	if data.ExpertRA != "" {
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "Эксперт РА",
			Rating:    data.ExpertRA,
			Score:     ratingToScore(data.ExpertRA),
		})
	}
	if data.Fitch != "" {
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "Fitch",
			Rating:    data.Fitch,
			Score:     ratingToScore(data.Fitch),
		})
	}
	if data.Moody != "" {
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "Moody's",
			Rating:    data.Moody,
			Score:     ratingToScore(data.Moody),
		})
	}
	if data.SP != "" {
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "S&P",
			Rating:    data.SP,
			Score:     ratingToScore(data.SP),
		})
	}

	// If dohod.ru provides its own composite credit_rating, use it as the score for agency ratings
	dohodScore := int(data.CreditRating)
	if dohodScore > 0 {
		for i := range ratings {
			ratings[i].Score = dohodScore
		}
	}

	// Always save ДОХОДЪ internal rating when credit_rating > 0 (even if no agency ratings exist)
	if dohodScore > 0 {
		ratingText := data.CreditRatingText
		if ratingText == "" {
			ratingText = data.EstimationRatingText
		}
		if ratingText == "" {
			ratingText = fmt.Sprintf("%d/10", dohodScore)
		}
		ratings = append(ratings, model.IssuerRating{
			EmitterID: emitterID,
			Issuer:    issuerName,
			Agency:    "ДОХОДЪ",
			Rating:    ratingText,
			Score:     dohodScore,
		})
	}

	for _, r := range ratings {
		if err := s.ratingRepo.Upsert(ctx, &r); err != nil {
			log.Printf("WARN: update rating for emitter %d %s (%s): %v", emitterID, issuerName, r.Agency, err)
		}
	}

	if len(ratings) > 0 {
		log.Printf("Updated %d ratings for emitter %d %q from dohod.ru (score=%d)", len(ratings), emitterID, issuerName, dohodScore)
	}
}

// SyncRatingsResult holds stats from a bulk rating sync run.
type SyncRatingsResult struct {
	TotalEmitters int
	AlreadyRated  int
	Processed     int
	NewlyRated    int
	Errors        int
}

// SyncAllRatings fetches dohod.ru data for one bond per emitter to populate credit ratings.
// Processes emitters sequentially with a delay to avoid hammering dohod.ru.
// If onlyMissing is true, skips emitters that already have ratings.
func (s *DetailsService) SyncAllRatings(ctx context.Context, onlyMissing bool, delayBetween time.Duration) (*SyncRatingsResult, error) {
	if s.issuerRepo == nil || s.ratingRepo == nil {
		return nil, fmt.Errorf("issuerRepo and ratingRepo are required for sync")
	}

	// Get one sample secid per emitter
	emitterSecids, err := s.issuerRepo.GetOneSecidPerEmitter(ctx)
	if err != nil {
		return nil, fmt.Errorf("get emitters: %w", err)
	}

	result := &SyncRatingsResult{TotalEmitters: len(emitterSecids)}

	// Get already-rated emitters
	var ratedSet map[int64]bool
	if onlyMissing {
		ratedSet, err = s.ratingRepo.GetDistinctEmitterIDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("get rated emitters: %w", err)
		}
		result.AlreadyRated = len(ratedSet)
	}

	for emitterID, secid := range emitterSecids {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		if onlyMissing && ratedSet[emitterID] {
			continue
		}

		// Look up ISIN from bond detail cache (redis) or MOEX
		issuer, err := s.issuerRepo.GetBySecid(ctx, secid)
		if err != nil || issuer == nil {
			log.Printf("[sync] skip emitter %d: cannot find secid %s", emitterID, secid)
			result.Errors++
			continue
		}

		// We need ISIN for dohod.ru. Try to get it from dohod cache first.
		cached, _ := s.detailsRepo.Get(ctx, secid)
		if cached == nil {
			cached, _ = s.detailsRepo.GetBySecid(ctx, secid)
		}
		isin := ""
		if cached != nil {
			isin = cached.ISIN
		}
		if isin == "" {
			// ISIN equals secid for most Russian bonds
			isin = secid
		}

		log.Printf("[sync] %d/%d — emitter %d, secid=%s, isin=%s", result.Processed+1, result.TotalEmitters-result.AlreadyRated, emitterID, secid, isin)

		_, fetchErr := s.FetchAndSave(ctx, secid, isin)
		if fetchErr != nil {
			log.Printf("[sync] ERROR emitter %d (%s): %v", emitterID, secid, fetchErr)
			result.Errors++
		} else {
			result.NewlyRated++
		}
		result.Processed++

		// Rate limit
		if delayBetween > 0 {
			select {
			case <-time.After(delayBetween):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}
	}

	return result, nil
}

// ratingToScore converts a rating string like "AAA(RU)", "ruAA+", "AA-" to a 1-10 score.
func ratingToScore(rating string) int {
	// Normalize: remove prefixes and suffixes like (RU), ru, (EXP)
	r := rating
	for _, prefix := range []string{"ru", "Ru", "RU"} {
		if len(r) > len(prefix) && r[:len(prefix)] == prefix {
			r = r[len(prefix):]
			break
		}
	}
	// Remove parenthetical suffixes like (RU), (EXP)
	for i := range r {
		if r[i] == '(' {
			r = r[:i]
			break
		}
	}

	switch r {
	case "AAA":
		return 10
	case "AA+":
		return 9
	case "AA":
		return 8
	case "AA-":
		return 7
	case "A+":
		return 6
	case "A":
		return 5
	case "A-":
		return 5
	case "BBB+":
		return 4
	case "BBB":
		return 4
	case "BBB-":
		return 3
	case "BB+":
		return 3
	case "BB":
		return 2
	case "BB-":
		return 2
	case "B+", "B", "B-":
		return 1
	default:
		return 0
	}
}
