package scoring

import (
	"math"
	"testing"
	"time"

	"nla/internal/model"
)

// floatPtr is a tiny ergonomic helper for the nullable bond fields.
func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int           { return &v }

// goodBond builds a bond that scores reasonably well across every factor.
// Tests start from goodBond and mutate one field to isolate behaviour.
func goodBond() model.Bond {
	return model.Bond{
		SECID:           "RU0000TEST",
		EffectiveYield:  floatPtr(12.0),
		Yield:           floatPtr(12.0),
		Duration:        intPtr(540), // ~1.5y
		ValueTodayRUB:   floatPtr(50_000_000),
		BondCategory:    "Корпоративная",
		DaysToPut:       nil,
		IssueSizePlaced: 5_000_000, // 5M lots
		FaceValue:       1000,      // 1000 RUB face → 5B placement
		IsFloat:         false,
		IsIndexed:       false,
	}
}

func goodRatings() []model.IssuerRating {
	return []model.IssuerRating{
		{Agency: "АКРА", Rating: "AA-(RU)", ScoreOrd: 19, UpdatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{Agency: "Эксперт РА", Rating: "ruA+", ScoreOrd: 18, UpdatedAt: time.Now().Add(-60 * 24 * time.Hour)},
	}
}

func goodDohod() *model.DohodBondData {
	q := 7.0
	s := 6.0
	return &model.DohodBondData{Quality: &q, Stability: &s}
}

// ---------- factor extractor tests ----------

func TestExtractCreditRating(t *testing.T) {
	raw, ok := extractCreditRating(Input{Ratings: goodRatings()})
	if !ok || raw != 19 {
		t.Fatalf("want best ord 19, got %v ok=%v", raw, ok)
	}

	if _, ok := extractCreditRating(Input{Ratings: nil}); ok {
		t.Fatal("empty ratings must yield hasData=false")
	}
	// All-zero ord → unrated → missing.
	zero := []model.IssuerRating{{Agency: "Foo", ScoreOrd: 0}}
	if _, ok := extractCreditRating(Input{Ratings: zero}); ok {
		t.Fatal("ord=0 must yield hasData=false")
	}
}

func TestExtractYTMPrefersEffective(t *testing.T) {
	b := goodBond()
	b.EffectiveYield = floatPtr(14.5)
	b.Yield = floatPtr(11)
	raw, ok := extractYTM(Input{Bond: b})
	if !ok || raw != 14.5 {
		t.Fatalf("expected effective_yield 14.5, got %v ok=%v", raw, ok)
	}

	b.EffectiveYield = nil
	raw, ok = extractYTM(Input{Bond: b})
	if !ok || raw != 11 {
		t.Fatalf("expected fallback yield 11, got %v ok=%v", raw, ok)
	}

	b.Yield = nil
	if _, ok := extractYTM(Input{Bond: b}); ok {
		t.Fatal("both nil → missing")
	}
}

func TestExtractYTMPremium(t *testing.T) {
	b := goodBond()
	bench := 10.0
	raw, ok := extractYTMPremium(Input{Bond: b, BenchmarkYieldPct: &bench})
	if !ok || raw != 2 { // ytm 12 - bench 10
		t.Fatalf("want premium=2, got %v ok=%v", raw, ok)
	}

	if _, ok := extractYTMPremium(Input{Bond: b}); ok {
		t.Fatal("missing benchmark → factor missing")
	}
}

func TestExtractPutOfferSoon(t *testing.T) {
	b := goodBond()
	b.DaysToPut = intPtr(30)
	raw, ok := extractPutOfferSoon(Input{Bond: b})
	if !ok || raw != 1 {
		t.Fatalf("put in 30d → 1.0, got %v ok=%v", raw, ok)
	}

	b.DaysToPut = intPtr(180)
	raw, _ = extractPutOfferSoon(Input{Bond: b})
	if raw != 0 {
		t.Fatalf("put in 180d → 0.0, got %v", raw)
	}

	b.DaysToPut = nil
	raw, ok = extractPutOfferSoon(Input{Bond: b})
	if !ok || raw != 0 {
		t.Fatalf("no put date → 0.0 with hasData=true, got %v ok=%v", raw, ok)
	}
}

func TestExtractCategoryVDOProxy(t *testing.T) {
	b := goodBond()
	b.BondCategory = "Корпоративная"
	b.Yield = floatPtr(22)
	raw, ok := extractCategory(Input{Bond: b})
	if !ok || raw != catVDOproxy {
		t.Fatalf("yield 22%% corp → VDO proxy, got %v ok=%v", raw, ok)
	}

	b.Yield = floatPtr(10)
	raw, _ = extractCategory(Input{Bond: b})
	if raw != catCorp {
		t.Fatalf("yield 10%% corp → corp, got %v", raw)
	}

	b.BondCategory = ""
	if _, ok := extractCategory(Input{Bond: b}); ok {
		t.Fatal("empty category → missing")
	}
}

func TestExtractCouponType(t *testing.T) {
	b := goodBond()
	raw, _ := extractCouponType(Input{Bond: b})
	if raw != couponFixed {
		t.Fatalf("fixed bond, got %v", raw)
	}

	b.IsFloat = true
	raw, _ = extractCouponType(Input{Bond: b})
	if raw != couponFloat {
		t.Fatalf("float, got %v", raw)
	}

	b.IsFloat = false
	b.IsIndexed = true
	raw, _ = extractCouponType(Input{Bond: b})
	if raw != couponIndexed {
		t.Fatalf("indexed, got %v", raw)
	}
}

func TestExtractRatingAge(t *testing.T) {
	ratings := []model.IssuerRating{
		{UpdatedAt: time.Now().Add(-100 * 24 * time.Hour)},
		{UpdatedAt: time.Now().Add(-30 * 24 * time.Hour)}, // freshest
		{UpdatedAt: time.Now().Add(-365 * 24 * time.Hour)},
	}
	raw, ok := extractRatingAge(Input{Ratings: ratings})
	if !ok {
		t.Fatal("expected hasData=true")
	}
	// freshest is 30d old; tolerate ±1d slop
	if raw < 29 || raw > 31 {
		t.Fatalf("expected ~30d, got %v", raw)
	}

	if _, ok := extractRatingAge(Input{Ratings: nil}); ok {
		t.Fatal("empty ratings → missing")
	}
}

func TestExtractIssueSize(t *testing.T) {
	b := goodBond()
	raw, ok := extractIssueSize(Input{Bond: b})
	if !ok || raw != 5_000_000_000 {
		t.Fatalf("5M lots * 1000 face = 5B, got %v ok=%v", raw, ok)
	}

	b.IssueSizePlaced = 0
	b.IssueSize = 1_000_000
	raw, ok = extractIssueSize(Input{Bond: b})
	if !ok || raw != 1_000_000_000 {
		t.Fatalf("fallback to issuesize: 1M * 1000 = 1B, got %v ok=%v", raw, ok)
	}

	b.IssueSize = 0
	if _, ok := extractIssueSize(Input{Bond: b}); ok {
		t.Fatal("no placed/issue size → missing")
	}
}

func TestExtractDohodFactors(t *testing.T) {
	d := goodDohod()
	if raw, ok := extractDohodQuality(Input{Dohod: d}); !ok || raw != 7 {
		t.Fatalf("quality 7, got %v ok=%v", raw, ok)
	}
	if raw, ok := extractDohodStability(Input{Dohod: d}); !ok || raw != 6 {
		t.Fatalf("stability 6, got %v ok=%v", raw, ok)
	}

	if _, ok := extractDohodQuality(Input{Dohod: nil}); ok {
		t.Fatal("nil dohod → missing")
	}
}

// ---------- normalize tests (anchor points) ----------

func TestNormalizeCreditRatingAnchors(t *testing.T) {
	cases := []struct {
		ord  float64
		want float64
	}{
		{1, 5},     // D
		{22, 100},  // AAA
		{14, 64.9}, // BBB-ish, ((14-1)/21)*95+5 ≈ 63.81; round generously
	}
	for _, c := range cases {
		got := normalizeCreditRating(c.ord)
		if got < c.want-2 || got > c.want+2 {
			t.Errorf("ord=%v want≈%v got %v", c.ord, c.want, got)
		}
	}
}

func TestNormalizeYTMSweetSpot(t *testing.T) {
	if n := normalizeYTM(17); n != 100 {
		t.Fatalf("17%% in sweet spot 15-20 → 100, got %v", n)
	}
	if n := normalizeYTM(30); n != 60 {
		t.Fatalf("30%% suspicious tier → 60, got %v", n)
	}
	if n := normalizeYTM(0); n != 0 {
		t.Fatalf("0%% → 0, got %v", n)
	}
	// Garbage MOEX data (>100%) must not score «high yield» — sink to 0.
	if n := normalizeYTM(2226487); n != 0 {
		t.Fatalf("bogus yield 2226487%% → 0, got %v", n)
	}
	if n := normalizeYTM(75); n != 15 {
		t.Fatalf("75%% junk tier → 15, got %v", n)
	}
}

func TestNormalizeMissingDataPolicy(t *testing.T) {
	// Positive-weight factor missing → neutral 50.
	if v := normalize(FactorCreditRating, 0, false, 0.4); v != 50 {
		t.Fatalf("positive weight + missing → 50, got %v", v)
	}
	// Negative-weight (penalty) missing → 0 (no penalty).
	if v := normalize(FactorPutOfferSoon, 0, false, -0.05); v != 0 {
		t.Fatalf("negative weight + missing → 0, got %v", v)
	}
}

// ---------- engine.Compute end-to-end ----------

func TestComputeGoodBondAcrossPresets(t *testing.T) {
	bench := 10.0
	in := Input{
		Bond:              goodBond(),
		Ratings:           goodRatings(),
		Dohod:             goodDohod(),
		BenchmarkYieldPct: &bench,
	}

	for _, p := range []Profile{PresetLow, PresetMid, PresetHigh} {
		res := Compute(in, p)
		if res.ProfileCode != p.Code {
			t.Errorf("profile code roundtrip broken: want %s got %s", p.Code, res.ProfileCode)
		}
		if len(res.Breakdown) != len(AllFactors) {
			t.Errorf("%s: breakdown should have %d items, got %d", p.Code, len(AllFactors), len(res.Breakdown))
		}
		if res.Score < 0 || res.Score > 100 {
			t.Errorf("%s: score must be clamped to [0,100], got %v", p.Code, res.Score)
		}
		if math.IsNaN(res.Score) {
			t.Errorf("%s: NaN score", p.Code)
		}
		// Score should be solidly above the neutral floor — this is a
		// "good" bond by construction.
		if res.Score < 40 {
			t.Errorf("%s: good bond should clear 40, got %v", p.Code, res.Score)
		}
	}
}

func TestComputePutOfferPenalisesLowOnly(t *testing.T) {
	bench := 10.0
	in := Input{
		Bond:              goodBond(),
		Ratings:           goodRatings(),
		BenchmarkYieldPct: &bench,
	}
	in.Bond.DaysToPut = intPtr(45)

	low := Compute(in, PresetLow)
	mid := Compute(in, PresetMid)
	high := Compute(in, PresetHigh)

	// Same bond, same data — put-offer penalty only bites under Low.
	if low.Score >= mid.Score {
		t.Errorf("put-offer in 45d should pull Low score below Mid; low=%v mid=%v", low.Score, mid.Score)
	}
	if low.Score >= high.Score {
		t.Errorf("put-offer in 45d should pull Low score below High; low=%v high=%v", low.Score, high.Score)
	}
}

func TestComputeMissingFactorsTracked(t *testing.T) {
	in := Input{
		Bond:    goodBond(),
		Ratings: nil, // no ratings → credit_rating + rating_age missing
		Dohod:   nil, // dohod_quality, dohod_stability missing (only matter for High)
	}
	// Benchmark intentionally nil → ytm_premium missing.

	res := Compute(in, PresetHigh)

	want := map[string]bool{
		FactorCreditRating:   true,
		FactorRatingAge:      true,
		FactorDohodQuality:   true,
		FactorDohodStability: true,
		FactorYTMPremium:     true,
	}
	for _, code := range res.MissingFactors {
		delete(want, code)
	}
	if len(want) > 0 {
		t.Errorf("missing-factors list incomplete, still expected: %v", want)
	}

	// Sanity: score still in range even with half the inputs gone.
	if res.Score < 0 || res.Score > 100 {
		t.Errorf("score out of range when half the factors missing: %v", res.Score)
	}
}

func TestComputeAllProfilesIncludeAllFactorsInBreakdown(t *testing.T) {
	in := Input{Bond: goodBond(), Ratings: goodRatings()}
	for _, p := range []Profile{PresetLow, PresetMid, PresetHigh} {
		res := Compute(in, p)
		got := map[string]bool{}
		for _, b := range res.Breakdown {
			got[b.Factor] = true
		}
		for _, code := range AllFactors {
			if !got[code] {
				t.Errorf("%s: factor %q missing from breakdown", p.Code, code)
			}
		}
	}
}

func TestPresetWeightsMatchRoadmap(t *testing.T) {
	// Lock in the weights so a future drift between presets.go and the DB
	// seed (migration 0002) is caught here.
	wantLow := map[string]float64{
		FactorCreditRating: 0.40, FactorYTM: 0.05, FactorYTMPremium: 0.05,
		FactorDuration: 0.15, FactorLiquidity: 0.15, FactorCategory: 0.05,
		FactorPutOfferSoon: -0.05, FactorIssueSize: 0.05, FactorCouponType: 0.05,
		FactorRatingAge: 0.05, FactorDohodQuality: 0, FactorDohodStability: 0,
	}
	for k, v := range wantLow {
		if PresetLow.Weights[k] != v {
			t.Errorf("PresetLow[%q] = %v, want %v", k, PresetLow.Weights[k], v)
		}
	}
}
