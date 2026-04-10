package dohod

import (
	"testing"
)

// Minimal Nuxt3 payload simulating analytics.dohod.ru response
var testPayloadCorpBond = []byte(`<html><head></head><body>
<script type="application/json" id="__NUXT_DATA__">[
  {"data":2,"state":33},
  ["ShallowReactive",3],
  {"sKey":4},
  {"TEST":5},
  "test_val",
  90,
  {"isin":7,"bond":8,"issuer":9,"borrower":13,"creditRating":14,"creditRatingText":15,"akra":16,"expert":17,"fitch":18,"moody":19,"sp":20,"estimationRating":21,"estimationRatingText":22,"quality":23,"qout":24,"qin":25,"qbalance":26,"qearnings":27,"qros":28,"qvalueROS":29,"qoperProf":30,"qdp1":31,"qearningsDP2":32,"qbalanceDP3":33,"qinventorTurnov":34,"qturnovOfCurAsset":35,"qreceivableTurnov":36,"qbalanceLiq":37,"qcurrentLiq":38,"qquiqLiq":39,"qstability":40,"qshortliabilities":41,"bestScore":42,"downRisk":43,"liquidity":44,"qratingROE":45,"qvalueROE":46,"qratingNetDebtEquity":47,"qvalueNetDebtEquity":48,"qvalueCashRatio":49,"qratingCashRatio":50,"qvalueCurrentLiq":51,"qvalueQuiqLiq":52,"qvalueShortliabilities":53,"qvalueOperProf":54,"qvalueInventorTurnov":55,"qvalueTurnovOfCurAsset":56,"qvalueReceivableTurnov":57,"totalReturn":58,"currentYield":59,"size":60,"complexity":61,"qcol":62},
  "RU000A106540",
  "bond_data_placeholder",
  {"nameShort":10,"economySector":11,"countryName":12},
  "ООО Тест Эмитент",
  "Финансы",
  "Россия",
  "Тестовый заёмщик",
  5.0,
  "BB-",
  "BBB+(RU)",
  "ruBBB",
  "",
  "",
  "",
  0,
  "Переоценка",
  2.55,
  2.14,
  2.95,
  1.13,
  4.63,
  1.0,
  -8.3,
  10.0,
  0.0,
  -0.51,
  -0.13,
  3.0,
  7.0,
  10.0,
  1.33,
  2.0,
  1.0,
  1.2,
  2.0,
  2.59,
  -0.6,
  48.97,
  1.0,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  null,
  0.0,
  12.02,
  10.0,
  0.0,
  5.5
]</script></body></html>`)

// Test payload for government bond (no agency ratings)
var testPayloadGovBond = []byte(`<html><head></head><body>
<script type="application/json" id="__NUXT_DATA__">[
  {"data":2,"state":10},
  ["ShallowReactive",3],
  {"sKey":4},
  {"TEST":5},
  "test_val",
  90,
  {"isin":7,"bond":8,"issuer":9,"creditRating":13,"creditRatingText":14,"quality":15,"akra":16,"expert":17,"fitch":18,"moody":19,"sp":20},
  "RU000A1038V6",
  "ofz_data",
  {"nameShort":10,"economySector":11,"countryName":12},
  "Минфин России",
  "орган власти РФ",
  "Россия",
  10.0,
  "AAA",
  10.0,
  "",
  "",
  "",
  "",
  ""
]</script></body></html>`)

func TestParseNuxtPayload_CorpBond(t *testing.T) {
	data, err := ParseNuxtPayload(testPayloadCorpBond, "RU000A106540")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.ISIN != "RU000A106540" {
		t.Errorf("ISIN = %q, want %q", data.ISIN, "RU000A106540")
	}

	// Issuer
	if data.IssuerName != "ООО Тест Эмитент" {
		t.Errorf("IssuerName = %q, want %q", data.IssuerName, "ООО Тест Эмитент")
	}
	if data.IssuerSector != "Финансы" {
		t.Errorf("IssuerSector = %q, want %q", data.IssuerSector, "Финансы")
	}
	if data.Country != "Россия" {
		t.Errorf("Country = %q, want %q", data.Country, "Россия")
	}

	// Credit ratings
	if data.CreditRating != 5.0 {
		t.Errorf("CreditRating = %v, want 5.0", data.CreditRating)
	}
	if data.CreditRatingText != "BB-" {
		t.Errorf("CreditRatingText = %q, want %q", data.CreditRatingText, "BB-")
	}
	if data.AKRA != "BBB+(RU)" {
		t.Errorf("AKRA = %q, want %q", data.AKRA, "BBB+(RU)")
	}
	if data.ExpertRA != "ruBBB" {
		t.Errorf("ExpertRA = %q, want %q", data.ExpertRA, "ruBBB")
	}

	// Quality
	assertFloatPtr(t, "Quality", data.Quality, 2.55)
	assertFloatPtr(t, "QualityOutside", data.QualityOutside, 2.14)
	assertFloatPtr(t, "QualityInside", data.QualityInside, 2.95)
	assertFloatPtr(t, "QualityBalance", data.QualityBalance, 1.13)
	assertFloatPtr(t, "QualityEarnings", data.QualityEarnings, 4.63)

	// DP
	assertFloatPtr(t, "DP1", data.DP1, 0.0)
	assertFloatPtr(t, "DP2", data.DP2, -0.51)
	assertFloatPtr(t, "DP3", data.DP3, -0.13)

	// Profitability
	assertFloatPtr(t, "ProfitROS", data.ProfitROS, 1.0)
	assertFloatPtr(t, "ProfitROSValue", data.ProfitROSValue, -8.3)
	assertFloatPtr(t, "ProfitOper", data.ProfitOper, 10.0)

	// Turnover
	assertFloatPtr(t, "TurnoverInventory", data.TurnoverInventory, 3.0)
	assertFloatPtr(t, "TurnoverCurAsset", data.TurnoverCurAsset, 7.0)
	assertFloatPtr(t, "TurnoverReceiv", data.TurnoverReceiv, 10.0)

	// Liquidity
	assertFloatPtr(t, "LiqBalance", data.LiqBalance, 1.33)
	assertFloatPtr(t, "LiqCurrent", data.LiqCurrent, 2.0)
	assertFloatPtr(t, "LiqQuick", data.LiqQuick, 1.0)

	// Stability
	assertFloatPtr(t, "Stability", data.Stability, 1.2)
	assertFloatPtr(t, "StabilityDebt", data.StabilityDebt, 2.0)

	// Metrics
	assertFloatPtr(t, "BestScore", data.BestScore, 2.59)
	assertFloatPtr(t, "DownRisk", data.DownRisk, -0.6)
	assertFloatPtr(t, "Liquidity", data.Liquidity, 48.97)

	// Current yield
	assertFloatPtr(t, "CurrentYield", data.CurrentYield, 12.02)

	// QualityProfitChg
	assertFloatPtr(t, "QualityProfitChg", data.QualityProfitChg, 5.5)
}

func TestParseNuxtPayload_GovBond(t *testing.T) {
	data, err := ParseNuxtPayload(testPayloadGovBond, "RU000A1038V6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.CreditRating != 10.0 {
		t.Errorf("CreditRating = %v, want 10.0", data.CreditRating)
	}
	if data.CreditRatingText != "AAA" {
		t.Errorf("CreditRatingText = %q, want %q", data.CreditRatingText, "AAA")
	}
	if data.IssuerName != "Минфин России" {
		t.Errorf("IssuerName = %q, want %q", data.IssuerName, "Минфин России")
	}

	// Agency ratings should be empty for gov bonds
	if data.AKRA != "" {
		t.Errorf("AKRA = %q, want empty", data.AKRA)
	}

	// Quality should be 10 for gov bonds
	assertFloatPtr(t, "Quality", data.Quality, 10.0)
}

func TestParseNuxtPayload_NoPayload(t *testing.T) {
	html := []byte(`<html><body>no payload here</body></html>`)
	_, err := ParseNuxtPayload(html, "RU000A1038V6")
	if err == nil {
		t.Fatal("expected error for missing payload")
	}
}

func TestParseNuxtPayload_NoBondObject(t *testing.T) {
	html := []byte(`<html><body><script type="application/json" id="__NUXT_DATA__">[1, 2, "hello", {"key": 1}]</script></body></html>`)
	_, err := ParseNuxtPayload(html, "RU000A1038V6")
	if err == nil {
		t.Fatal("expected error for missing bond object")
	}
}

func TestIsValidISIN(t *testing.T) {
	tests := []struct {
		isin  string
		valid bool
	}{
		{"RU000A1038V6", true},
		{"RU000A106540", true},
		{"US0378331005", true},
		{"", false},
		{"SHORT", false},
		{"RU000A1038v6", false},  // lowercase
		{"RU000A1038 6", false},  // space
		{"RU000A1038V6X", false}, // too long
	}

	for _, tt := range tests {
		if got := isValidISIN(tt.isin); got != tt.valid {
			t.Errorf("isValidISIN(%q) = %v, want %v", tt.isin, got, tt.valid)
		}
	}
}

func TestResolveNuxtRefs_DepthLimit(t *testing.T) {
	// Create a self-referencing payload — should not infinite loop
	payload := []any{
		map[string]any{"a": float64(0)}, // references itself
	}
	result := resolveNuxtRefs(payload, 0, 0)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// --- helpers ---

func assertFloatPtr(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %v", name, want)
		return
	}
	// Allow small delta for float comparison
	if diff := *got - want; diff > 0.01 || diff < -0.01 {
		t.Errorf("%s = %v, want %v", name, *got, want)
	}
}
