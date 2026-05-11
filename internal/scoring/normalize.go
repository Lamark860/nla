package scoring

// normalize maps a raw factor value onto the 0..100 scale used by the engine,
// where 100 = «best contribution» under the bond-investor mental model.
// The mapping is intentionally table-/piecewise-based (no statistical
// fitting) so weights stay interpretable and calibration happens by editing
// these tables, not by re-tuning a model.
//
// Missing data policy:
//   - Positive-weight factors with hasData=false → neutral 50, so an emitter
//     without that data isn't unfairly punished or rewarded.
//   - Negative-weight (penalty) factors with hasData=false → 0, so a missing
//     value doesn't trigger the penalty.
func normalize(code string, raw float64, hasData bool, weight float64) float64 {
	if !hasData {
		if weight < 0 {
			return 0
		}
		return 50
	}

	switch code {
	case FactorCreditRating:
		return normalizeCreditRating(raw)
	case FactorYTM:
		return normalizeYTM(raw)
	case FactorYTMPremium:
		return normalizeYTMPremium(raw)
	case FactorDuration:
		return normalizeDuration(raw)
	case FactorLiquidity:
		return normalizeLiquidity(raw)
	case FactorCategory:
		return normalizeCategory(raw)
	case FactorPutOfferSoon:
		return raw * 100 // 0 or 1 → 0 or 100
	case FactorIssueSize:
		return normalizeIssueSize(raw)
	case FactorCouponType:
		return normalizeCouponType(raw)
	case FactorRatingAge:
		return normalizeRatingAge(raw)
	case FactorDohodQuality, FactorDohodStability:
		return clamp(raw*10, 0, 100) // dohod scale 0..10 → 0..100
	}
	return 50
}

// normalizeCreditRating: 22-level ord onto 0..100. ord=1 (D) → 5,
// ord=22 (AAA) → 100. Linear in between.
func normalizeCreditRating(ord float64) float64 {
	if ord < 1 {
		return 0
	}
	if ord > 22 {
		ord = 22
	}
	return clamp(((ord-1)/21)*95+5, 0, 100)
}

// normalizeYTM: piecewise. Very low YTM (<5%) is below риск-free → 20;
// very high YTM (>25%) is suspicious / distressed → 60. Sweet spot 15-20%
// gets the max.
func normalizeYTM(y float64) float64 {
	switch {
	case y <= 0:
		return 0
	case y < 5:
		return 20
	case y < 10:
		return 40
	case y < 15:
		return 70
	case y < 20:
		return 100
	case y < 25:
		return 85
	default:
		return 60
	}
}

// normalizeYTMPremium: relative spread vs benchmark in %. Negative spreads
// (below ОФЗ) → low score; +1..+3 pp → solid; >+5 pp likely distressed.
func normalizeYTMPremium(p float64) float64 {
	switch {
	case p < -1:
		return 20
	case p < 0:
		return 35
	case p < 1:
		return 55
	case p < 3:
		return 80
	case p < 5:
		return 100
	case p < 8:
		return 85
	default:
		return 60
	}
}

// normalizeDuration: shorter = better for the Low-risk lens; weights flip
// the contribution for High. Returns 100 for <6 mo, 20 for >5 yr.
func normalizeDuration(days float64) float64 {
	switch {
	case days <= 180:
		return 100
	case days <= 365:
		return 85
	case days <= 730:
		return 70
	case days <= 1100:
		return 55
	case days <= 1825:
		return 40
	default:
		return 20
	}
}

// normalizeLiquidity: RUB turnover today. Log-ish buckets — 100M+ → 100,
// <100k → 10.
func normalizeLiquidity(rub float64) float64 {
	switch {
	case rub < 100_000:
		return 10
	case rub < 1_000_000:
		return 30
	case rub < 10_000_000:
		return 55
	case rub < 100_000_000:
		return 80
	default:
		return 100
	}
}

// normalizeCategory: ОФЗ → safest, Корп → mid, ВДО proxy → low.
func normalizeCategory(cat float64) float64 {
	switch cat {
	case catOFZ:
		return 100
	case catSubfed:
		return 80
	case catMuni:
		return 70
	case catMortgage:
		return 70
	case catCorp:
		return 60
	case catEurobond:
		return 50
	case catVDOproxy:
		return 35
	default:
		return 40
	}
}

// normalizeIssueSize: bigger placement = better liquidity story.
// <100M RUB → 20, >10B → 100.
func normalizeIssueSize(rub float64) float64 {
	switch {
	case rub < 100_000_000:
		return 20
	case rub < 1_000_000_000:
		return 50
	case rub < 10_000_000_000:
		return 75
	default:
		return 100
	}
}

// normalizeCouponType: fixed = predictable cash flow = best.
func normalizeCouponType(t float64) float64 {
	switch t {
	case couponFixed:
		return 100
	case couponIndexed:
		return 70
	case couponFloat:
		return 50
	default:
		return 50
	}
}

// normalizeRatingAge: penalise stale ratings. <90d fresh → 100, ≥2y stale → 20.
func normalizeRatingAge(days float64) float64 {
	switch {
	case days < 90:
		return 100
	case days < 180:
		return 85
	case days < 365:
		return 65
	case days < 730:
		return 40
	default:
		return 20
	}
}
