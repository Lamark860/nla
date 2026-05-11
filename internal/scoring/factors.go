package scoring

import (
	"time"
)

// extractFactor dispatches to the per-factor extractor. Each extractor returns
// the raw value in its native unit (ord, % p.a., days, RUB, …) plus a flag
// telling whether the data was present at all. Missing data → the normalizer
// decides what to substitute (neutral 50 for positive-weight factors,
// 0 for penalties).
func extractFactor(code string, in Input) (raw float64, hasData bool) {
	switch code {
	case FactorCreditRating:
		return extractCreditRating(in)
	case FactorYTM:
		return extractYTM(in)
	case FactorYTMPremium:
		return extractYTMPremium(in)
	case FactorDuration:
		return extractDuration(in)
	case FactorLiquidity:
		return extractLiquidity(in)
	case FactorCategory:
		return extractCategory(in)
	case FactorPutOfferSoon:
		return extractPutOfferSoon(in)
	case FactorIssueSize:
		return extractIssueSize(in)
	case FactorCouponType:
		return extractCouponType(in)
	case FactorRatingAge:
		return extractRatingAge(in)
	case FactorDohodQuality:
		return extractDohodQuality(in)
	case FactorDohodStability:
		return extractDohodStability(in)
	}
	return 0, false
}

// 1. Best (highest) score_ord across the issuer's agency ratings. 1..22.
func extractCreditRating(in Input) (float64, bool) {
	best := 0
	for _, r := range in.Ratings {
		if r.ScoreOrd > best {
			best = r.ScoreOrd
		}
	}
	if best == 0 {
		return 0, false
	}
	return float64(best), true
}

// 2. Yield to maturity in % per year. Prefer EffectiveYield, fall back to
// the marketdata Yield field.
func extractYTM(in Input) (float64, bool) {
	if y := in.Bond.EffectiveYield; y != nil && *y > 0 {
		return *y, true
	}
	if y := in.Bond.Yield; y != nil && *y > 0 {
		return *y, true
	}
	return 0, false
}

// 3. YTM minus an externally provided ОФЗ benchmark of the same duration.
// Both values are in % p.a. Result can be negative.
func extractYTMPremium(in Input) (float64, bool) {
	if in.BenchmarkYieldPct == nil {
		return 0, false
	}
	ytm, ok := extractYTM(in)
	if !ok {
		return 0, false
	}
	return ytm - *in.BenchmarkYieldPct, true
}

// 4. Duration in days (MOEX reports duration as integer days).
func extractDuration(in Input) (float64, bool) {
	if in.Bond.Duration == nil || *in.Bond.Duration <= 0 {
		return 0, false
	}
	return float64(*in.Bond.Duration), true
}

// 5. Liquidity proxy — today's RUB turnover. A 30-day average would be
// nicer but requires a history aggregate we don't have at call sites yet;
// see roadmap Open tech debt.
func extractLiquidity(in Input) (float64, bool) {
	if in.Bond.ValueTodayRUB != nil && *in.Bond.ValueTodayRUB > 0 {
		return *in.Bond.ValueTodayRUB, true
	}
	if v := in.Bond.ValToday; v != nil && *v > 0 {
		return *v, true
	}
	return 0, false
}

// Numeric encoding of the BondCategory string. Used both for extraction
// (so engine treats categories uniformly) and as the lookup key in
// normalize.go.
const (
	catOFZ       = 1.0
	catSubfed    = 2.0
	catMuni      = 3.0
	catCorp      = 4.0
	catMortgage  = 5.0
	catEurobond  = 6.0
	catVDOproxy  = 7.0
	catUnknownID = 0.0
)

func extractCategory(in Input) (float64, bool) {
	switch in.Bond.BondCategory {
	case "ОФЗ":
		return catOFZ, true
	case "Субфедеральная":
		return catSubfed, true
	case "Муниципальная":
		return catMuni, true
	case "Корпоративная":
		// «ВДО» proxy: corporate + very high yield → treat as риск-категория
		if in.Bond.Yield != nil && *in.Bond.Yield >= 18 {
			return catVDOproxy, true
		}
		return catCorp, true
	case "Ипотечная":
		return catMortgage, true
	case "Еврооблигация":
		return catEurobond, true
	case "":
		return catUnknownID, false
	default:
		return catUnknownID, false
	}
}

// 7. PUT offer flag. Returns 1.0 if a put offer is within the next 90 days
// (i.e. holders can be assigned early redemption), 0.0 otherwise. Always
// hasData=true — absence of a put date means there is no near put.
func extractPutOfferSoon(in Input) (float64, bool) {
	if in.Bond.DaysToPut != nil && *in.Bond.DaysToPut > 0 && *in.Bond.DaysToPut < 90 {
		return 1.0, true
	}
	return 0.0, true
}

// 8. Issue size in RUB = placed lots × face value.
func extractIssueSize(in Input) (float64, bool) {
	placed := in.Bond.IssueSizePlaced
	if placed <= 0 {
		placed = in.Bond.IssueSize
	}
	if placed <= 0 || in.Bond.FaceValue <= 0 {
		return 0, false
	}
	return float64(placed) * in.Bond.FaceValue, true
}

// Coupon-type encoding. Used both as the extractor output and as the
// normalize key.
const (
	couponFixed   = 1.0
	couponIndexed = 2.0
	couponFloat   = 3.0
)

// 9. Coupon type: fixed = best, indexed = neutral, float = worst (for
// Low-risk profile; weights pick the contribution).
func extractCouponType(in Input) (float64, bool) {
	switch {
	case in.Bond.IsFloat:
		return couponFloat, true
	case in.Bond.IsIndexed:
		return couponIndexed, true
	default:
		return couponFixed, true
	}
}

// 10. Age of the freshest rating in days.
func extractRatingAge(in Input) (float64, bool) {
	var newest time.Time
	for _, r := range in.Ratings {
		if r.UpdatedAt.After(newest) {
			newest = r.UpdatedAt
		}
	}
	if newest.IsZero() {
		return 0, false
	}
	days := time.Since(newest).Hours() / 24
	if days < 0 {
		days = 0
	}
	return days, true
}

// 11. dohod.ru aggregate quality score (0..10 by their scale).
func extractDohodQuality(in Input) (float64, bool) {
	if in.Dohod == nil || in.Dohod.Quality == nil {
		return 0, false
	}
	return *in.Dohod.Quality, true
}

// 12. dohod.ru aggregate stability score (0..10 by their scale).
func extractDohodStability(in Input) (float64, bool) {
	if in.Dohod == nil || in.Dohod.Stability == nil {
		return 0, false
	}
	return *in.Dohod.Stability, true
}
