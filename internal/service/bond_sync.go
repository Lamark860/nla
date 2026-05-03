package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"nla/internal/model"
)

// SyncMissingIssuers finds bonds present in MOEX but absent from bond_issuers,
// fetches their EMITTER_ID from MOEX description API, and creates bond_issuer records.
func (s *BondService) SyncMissingIssuers(ctx context.Context) (int, error) {
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return 0, fmt.Errorf("get bonds: %w", err)
	}

	existing, err := s.issuerRepo.GetAllSecids(ctx)
	if err != nil {
		return 0, fmt.Errorf("get existing secids: %w", err)
	}

	var missing []string
	for _, b := range bonds {
		if !existing[b.SECID] {
			missing = append(missing, b.SECID)
		}
	}

	if len(missing) == 0 {
		return 0, nil
	}

	log.Printf("[issuer-sync] Found %d bonds missing from bond_issuers", len(missing))

	synced := 0
	for _, secid := range missing {
		select {
		case <-ctx.Done():
			return synced, ctx.Err()
		default:
		}

		raw, err := s.moex.GetDisclosure(ctx, secid)
		if err != nil {
			log.Printf("[issuer-sync] WARN: fetch disclosure for %s: %v", secid, err)
			continue
		}

		descRows := extractRows(raw, "description")
		var emitterID int64
		for _, row := range descRows {
			if getString(row, "name") == "EMITTER_ID" {
				if v, err := strconv.ParseInt(getString(row, "value"), 10, 64); err == nil {
					emitterID = v
				}
				break
			}
		}

		issuer := &model.BondIssuer{
			SECID:     secid,
			EmitterID: emitterID,
		}
		if err := s.issuerRepo.Upsert(ctx, issuer); err != nil {
			log.Printf("[issuer-sync] WARN: upsert issuer %s: %v", secid, err)
			continue
		}
		synced++
		log.Printf("[issuer-sync] Added %s (emitter_id=%d)", secid, emitterID)

		// Small delay to avoid hammering MOEX
		time.Sleep(200 * time.Millisecond)
	}

	return synced, nil
}

// SyncMissingRatingsFromMoex fetches credit ratings from MOEX CCI API
// for emitters that exist in bond_issuers but have no records in issuer_ratings.
func (s *BondService) SyncMissingRatingsFromMoex(ctx context.Context) (int, error) {
	allIssuers, err := s.issuerRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("get issuers: %w", err)
	}

	rated, err := s.ratingRepo.GetDistinctEmitterIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("get rated emitters: %w", err)
	}

	type missingInfo struct {
		emitterID   int64
		emitterName string
	}
	seen := make(map[int64]bool)
	var missing []missingInfo
	for _, iss := range allIssuers {
		if iss.EmitterID == 0 || rated[iss.EmitterID] || seen[iss.EmitterID] {
			continue
		}
		seen[iss.EmitterID] = true
		missing = append(missing, missingInfo{emitterID: iss.EmitterID, emitterName: iss.EmitterName})
	}

	if len(missing) == 0 {
		return 0, nil
	}

	log.Printf("[rating-sync] Found %d emitters without ratings, trying MOEX CCI", len(missing))

	synced := 0
	for _, m := range missing {
		select {
		case <-ctx.Done():
			return synced, ctx.Err()
		default:
		}

		cciRatings, err := s.moex.GetCCIRatings(ctx, m.emitterID)
		if err != nil {
			log.Printf("[rating-sync] WARN: CCI fetch for emitter %d: %v", m.emitterID, err)
			continue
		}
		if len(cciRatings) == 0 {
			continue
		}

		for _, cr := range cciRatings {
			rating := &model.IssuerRating{
				EmitterID: m.emitterID,
				Issuer:    m.emitterName,
				Agency:    cr.AgencyName,
				Rating:    cr.RatingValue,
			}
			if err := s.ratingRepo.Upsert(ctx, rating); err != nil {
				log.Printf("[rating-sync] WARN: upsert rating emitter %d %s: %v", m.emitterID, cr.AgencyName, err)
			}
		}

		synced++
		log.Printf("[rating-sync] MOEX CCI: emitter %d %q → %d ratings", m.emitterID, m.emitterName, len(cciRatings))

		time.Sleep(200 * time.Millisecond)
	}

	return synced, nil
}
