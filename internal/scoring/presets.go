package scoring

// Preset profile codes. Mirror the seed rows in
// internal/database/migrations/0002_postgres_full_schema.sql.
const (
	ProfileLow  = "low"
	ProfileMid  = "mid"
	ProfileHigh = "high"
)

// PresetLow — «Низкий риск» 🛡️.
//
// Heavy on credit rating + short duration + liquidity. Small PUT-offer
// penalty so an early-callable bond gets dinged in this profile only.
var PresetLow = Profile{
	Code: ProfileLow,
	Name: "Низкий риск",
	Weights: map[string]float64{
		FactorCreditRating:   0.40,
		FactorYTM:            0.05,
		FactorYTMPremium:     0.05,
		FactorDuration:       0.15,
		FactorLiquidity:      0.15,
		FactorCategory:       0.05,
		FactorPutOfferSoon:   -0.05,
		FactorIssueSize:      0.05,
		FactorCouponType:     0.05,
		FactorRatingAge:      0.05,
		FactorDohodQuality:   0.00,
		FactorDohodStability: 0.00,
	},
}

// PresetMid — «Средний риск» ⚖️.
//
// Balanced: rating still matters but yield and benchmark premium pull
// equal weight.
var PresetMid = Profile{
	Code: ProfileMid,
	Name: "Средний риск",
	Weights: map[string]float64{
		FactorCreditRating:   0.25,
		FactorYTM:            0.20,
		FactorYTMPremium:     0.15,
		FactorDuration:       0.10,
		FactorLiquidity:      0.10,
		FactorCategory:       0.05,
		FactorPutOfferSoon:   0.00,
		FactorIssueSize:      0.05,
		FactorCouponType:     0.05,
		FactorRatingAge:      0.05,
		FactorDohodQuality:   0.00,
		FactorDohodStability: 0.00,
	},
}

// PresetHigh — «Повышенный риск» 🚀.
//
// Yield-led with dohod.ru quality + stability factors switched on
// (they're the only signal of fundamental issuer health for ВДО-class
// names).
var PresetHigh = Profile{
	Code: ProfileHigh,
	Name: "Повышенный риск",
	Weights: map[string]float64{
		FactorCreditRating:   0.10,
		FactorYTM:            0.30,
		FactorYTMPremium:     0.25,
		FactorDuration:       0.05,
		FactorLiquidity:      0.05,
		FactorCategory:       0.05,
		FactorPutOfferSoon:   0.00,
		FactorIssueSize:      0.05,
		FactorCouponType:     0.05,
		FactorRatingAge:      0.05,
		FactorDohodQuality:   0.05,
		FactorDohodStability: 0.05,
	},
}

// Presets exposes the three preset profiles by code.
var Presets = map[string]Profile{
	ProfileLow:  PresetLow,
	ProfileMid:  PresetMid,
	ProfileHigh: PresetHigh,
}
