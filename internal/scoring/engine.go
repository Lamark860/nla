// Package scoring computes the deterministic «Аналитический индекс» (0..100)
// for a MOEX bond, given a weight profile (Low / Mid / High risk).
//
// The pipeline is:
//
//	Input{Bond, Ratings, Dohod, Bench} ──▶ 12 factor extractors
//	         (factors.go)                  → (raw, hasData)
//	                                       ──▶ normalize to 0..100
//	                                              (normalize.go)
//	                                       ──▶ weighted sum + clamp
//	                                              (this file)
//
// Profile.Weights are float64 keyed by the canonical factor codes declared
// below. The same keys live as JSONB columns in `scoring_profiles.weights`,
// so weights round-trip between the DB and this package without translation.
package scoring

import (
	"math"

	"nla/internal/model"
)

// Canonical factor codes. Match `scoring_profiles.weights` JSONB keys.
const (
	FactorCreditRating   = "credit_rating"
	FactorYTM            = "ytm"
	FactorYTMPremium     = "ytm_premium"
	FactorDuration       = "duration"
	FactorLiquidity      = "liquidity"
	FactorCategory       = "category"
	FactorPutOfferSoon   = "put_offer_soon"
	FactorIssueSize      = "issue_size"
	FactorCouponType     = "coupon_type"
	FactorRatingAge      = "rating_age"
	FactorDohodQuality   = "dohod_quality"
	FactorDohodStability = "dohod_stability"
)

// AllFactors lists factor codes in canonical breakdown order. Driven by
// docs/roadmap.md Фаза 2 — keep these synchronised.
var AllFactors = []string{
	FactorCreditRating,
	FactorYTM,
	FactorYTMPremium,
	FactorDuration,
	FactorLiquidity,
	FactorCategory,
	FactorPutOfferSoon,
	FactorIssueSize,
	FactorCouponType,
	FactorRatingAge,
	FactorDohodQuality,
	FactorDohodStability,
}

// factorNamesRU is the human-readable RU label per factor — surfaced in
// breakdown JSON for the UI.
var factorNamesRU = map[string]string{
	FactorCreditRating:   "Кредитный рейтинг",
	FactorYTM:            "Доходность к погашению",
	FactorYTMPremium:     "Премия к ОФЗ",
	FactorDuration:       "Дюрация",
	FactorLiquidity:      "Ликвидность",
	FactorCategory:       "Категория выпуска",
	FactorPutOfferSoon:   "Близкая PUT-оферта",
	FactorIssueSize:      "Размер эмиссии",
	FactorCouponType:     "Тип купона",
	FactorRatingAge:      "Свежесть рейтинга",
	FactorDohodQuality:   "Качество (dohod)",
	FactorDohodStability: "Стабильность (dohod)",
}

// Input bundles everything the engine needs to score one bond. Ratings can
// hold zero or more agency entries for the same emitter; the engine picks the
// strongest by ScoreOrd. Dohod and Benchmark may be nil — the corresponding
// factors will be marked HasData=false.
type Input struct {
	Bond    model.Bond
	Ratings []model.IssuerRating
	Dohod   *model.DohodBondData

	// BenchmarkYieldPct is the ОФЗ yield for the bond's duration bucket,
	// in % p.a. Used by FactorYTMPremium. Nil → factor missing.
	BenchmarkYieldPct *float64
}

// Profile is a named weight set. Loaded from `scoring_profiles` for presets
// and custom user profiles. Sum of weights typically ≈ 1.0; the engine does
// not enforce it.
type Profile struct {
	Code    string
	Name    string
	Weights map[string]float64
}

// BreakdownItem captures one factor's contribution to the final score, both
// for transparency in the UI and for caching in `bond_scores.breakdown`.
type BreakdownItem struct {
	Factor       string   `json:"factor"`
	Name         string   `json:"name"`
	Raw          *float64 `json:"raw,omitempty"`
	Normalized   float64  `json:"normalized"`
	Weight       float64  `json:"weight"`
	Contribution float64  `json:"contribution"`
	HasData      bool     `json:"has_data"`
}

// ScoreResult is what Compute returns and what the API hands to the frontend.
// Score is clamped to [0,100].
type ScoreResult struct {
	Score          float64         `json:"score"`
	ProfileCode    string          `json:"profile_code"`
	Breakdown      []BreakdownItem `json:"breakdown"`
	MissingFactors []string        `json:"missing_factors,omitempty"`
}

// Compute runs the scoring pipeline. It is pure: same input + same profile →
// same output. Designed to be called per-bond on demand, then cached in
// `bond_scores` for the day.
func Compute(in Input, profile Profile) ScoreResult {
	res := ScoreResult{
		ProfileCode: profile.Code,
		Breakdown:   make([]BreakdownItem, 0, len(AllFactors)),
	}

	for _, code := range AllFactors {
		raw, hasData := extractFactor(code, in)
		weight := profile.Weights[code]

		normalized := normalize(code, raw, hasData, weight)

		item := BreakdownItem{
			Factor:       code,
			Name:         factorNamesRU[code],
			Normalized:   normalized,
			Weight:       weight,
			Contribution: weight * normalized,
			HasData:      hasData,
		}
		if hasData {
			r := raw
			item.Raw = &r
		}

		res.Breakdown = append(res.Breakdown, item)
		res.Score += item.Contribution

		if !hasData && weight != 0 {
			res.MissingFactors = append(res.MissingFactors, code)
		}
	}

	res.Score = clamp(res.Score, 0, 100)
	return res
}

func clamp(v, lo, hi float64) float64 {
	if math.IsNaN(v) {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
