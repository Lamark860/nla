package service

import (
	"strconv"
	"testing"
)

func TestNormalizeRating(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantOrd int
		wantTier string
	}{
		// АКРА — "(RU)" suffix
		{"АКРА AAA", "AAA(RU)", 22, "AAA"},
		{"АКРА AA+", "AA+(RU)", 21, "AA"},
		{"АКРА AA", "AA(RU)", 20, "AA"},
		{"АКРА AA-", "AA-(RU)", 19, "AA"},
		{"АКРА A+", "A+(RU)", 18, "A"},
		{"АКРА A", "A(RU)", 17, "A"},
		{"АКРА A-", "A-(RU)", 16, "A"},
		{"АКРА BBB+", "BBB+(RU)", 15, "BBB"},
		{"АКРА BBB", "BBB(RU)", 14, "BBB"},
		{"АКРА BBB-", "BBB-(RU)", 13, "BBB"},
		{"АКРА BB+", "BB+(RU)", 12, "BB"},
		{"АКРА BB", "BB(RU)", 11, "BB"},
		{"АКРА BB-", "BB-(RU)", 10, "BB"},
		{"АКРА B+", "B+(RU)", 9, "B"},
		{"АКРА B", "B(RU)", 8, "B"},
		{"АКРА D", "D(RU)", 1, "D"},

		// Эксперт РА — "ru" prefix
		{"Эксперт ruAAA", "ruAAA", 22, "AAA"},
		{"Эксперт ruAA+", "ruAA+", 21, "AA"},
		{"Эксперт ruAA-", "ruAA-", 19, "AA"},
		{"Эксперт ruA+", "ruA+", 18, "A"},
		{"Эксперт ruBBB-", "ruBBB-", 13, "BBB"},
		{"Эксперт ruBB+", "ruBB+", 12, "BB"},
		{"Эксперт ruB", "ruB", 8, "B"},
		{"Эксперт ruC", "ruC", 2, "C"},
		{"Эксперт ruD", "ruD", 1, "D"},

		// НКР — ".ru" suffix
		{"НКР AAA.ru", "AAA.ru", 22, "AAA"},
		{"НКР BBB+.ru", "BBB+.ru", 15, "BBB"},
		{"НКР BB-.ru", "BB-.ru", 10, "BB"},
		{"НКР B.ru", "B.ru", 8, "B"},

		// НРА — "|ru|" suffix
		{"НРА AAA|ru|", "AAA|ru|", 22, "AAA"},
		{"НРА AA+|ru|", "AA+|ru|", 21, "AA"},
		{"НРА BBB|ru|", "BBB|ru|", 14, "BBB"},
		{"НРА BB|ru|", "BB|ru|", 11, "BB"},

		// S&P / Fitch — bare letters
		{"S&P AAA", "AAA", 22, "AAA"},
		{"Fitch AA-", "AA-", 19, "AA"},
		{"Fitch BBB+", "BBB+", 15, "BBB"},
		{"S&P BB", "BB", 11, "BB"},
		{"S&P D", "D", 1, "D"},
		{"Fitch RD", "RD", 1, "D"},

		// Moody's — case-sensitive notation
		{"Moody's Aaa", "Aaa", 22, "AAA"},
		{"Moody's Aa1", "Aa1", 21, "AA"},
		{"Moody's Aa2", "Aa2", 20, "AA"},
		{"Moody's Aa3", "Aa3", 19, "AA"},
		{"Moody's A1", "A1", 18, "A"},
		{"Moody's A2", "A2", 17, "A"},
		{"Moody's A3", "A3", 16, "A"},
		{"Moody's Baa1", "Baa1", 15, "BBB"},
		{"Moody's Baa2", "Baa2", 14, "BBB"},
		{"Moody's Baa3", "Baa3", 13, "BBB"},
		{"Moody's Ba1", "Ba1", 12, "BB"},
		{"Moody's Ba3", "Ba3", 10, "BB"},
		{"Moody's B2", "B2", 8, "B"},
		{"Moody's Caa1", "Caa1", 6, "CCC"},
		{"Moody's Ca", "Ca", 3, "CC"},
		{"Moody's C", "C", 2, "C"},

		// ДОХОДЪ — numeric 1-10 and "X/10" forms
		{"ДОХОДЪ 10", "10", 22, "AAA"},
		{"ДОХОДЪ 10/10", "10/10", 22, "AAA"},
		{"ДОХОДЪ 8", "8", 20, "AA"},
		{"ДОХОДЪ 7/10", "7/10", 19, "AA"},
		{"ДОХОДЪ 5", "5", 17, "A"},
		{"ДОХОДЪ 4", "4", 15, "BBB"},
		{"ДОХОДЪ 1", "1", 9, "B"},
		{"ДОХОДЪ 0 invalid", "0", 0, ""},
		{"ДОХОДЪ 11 invalid", "11", 0, ""},

		// Outlook stripping
		{"АКРА with outlook", "AAA(RU) Стабильный", 22, "AAA"},
		{"АКРА AA Negative parens", "AA(RU) (Negative)", 20, "AA"},
		{"Fitch BBB+ Stable", "BBB+ Stable", 15, "BBB"},
		{"S&P AA- Positive", "AA- Positive", 19, "AA"},
		{"Эксперт ruA+ развивающийся", "ruA+ развивающийся", 18, "A"},

		// Whitespace / case noise
		{"Whitespace", "  AAA  ", 22, "AAA"},
		{"Lower-case acronyms", "aaa(ru)", 22, "AAA"},
		{"Mixed case NKR", "BBB+.RU", 15, "BBB"},

		// Withdrawn / unrated
		{"Withdrawn ru", "Отозван", 0, ""},
		{"Withdrawn lat", "WD", 0, ""},
		{"Empty", "", 0, ""},
		{"Dash", "—", 0, ""},

		// Garbage in
		{"Bogus", "XYZ", 0, ""},
		{"Number out of range", "42", 0, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ord, tier := NormalizeRating(tc.input)
			if ord != tc.wantOrd {
				t.Errorf("NormalizeRating(%q): ord = %d, want %d", tc.input, ord, tc.wantOrd)
			}
			if tier != tc.wantTier {
				t.Errorf("NormalizeRating(%q): tier = %q, want %q", tc.input, tier, tc.wantTier)
			}
		})
	}
}

// Critical bug we are fixing: BBB- (investment-grade floor) must rank strictly
// higher than BB+ (top of speculative). The legacy ratingToScore function
// returned 3 for both, blurring the most important boundary in credit risk.
func TestNormalizeRating_InvestmentVsSpeculativeBoundary(t *testing.T) {
	bbbMinus, _ := NormalizeRating("BBB-(RU)")
	bbPlus, _ := NormalizeRating("BB+(RU)")
	if !(bbbMinus > bbPlus) {
		t.Fatalf("BBB- (%d) must rank above BB+ (%d)", bbbMinus, bbPlus)
	}
}

// Cross-agency ordering must match: an АКРА AAA must rank above an Эксперт РА A,
// regardless of which agency labelled it.
func TestNormalizeRating_CrossAgencyOrdering(t *testing.T) {
	pairs := []struct {
		stronger, weaker string
	}{
		{"AAA(RU)", "ruA"},          // АКРА AAA > Эксперт A
		{"AA+.ru", "BBB-(RU)"},      // НКР AA+ > АКРА BBB-
		{"Aaa", "ruBB+"},            // Moody's Aaa > Эксперт BB+
		{"BBB+|ru|", "ruBB-"},       // НРА BBB+ > Эксперт BB-
		{"AA-", "B+(RU)"},           // S&P AA- > АКРА B+
		{"Baa1", "Caa1"},            // Moody's Baa1 > Caa1
	}
	for _, p := range pairs {
		s, _ := NormalizeRating(p.stronger)
		w, _ := NormalizeRating(p.weaker)
		if !(s > w) {
			t.Errorf("expected %s (%d) > %s (%d)", p.stronger, s, p.weaker, w)
		}
	}
}

func TestLegacyScore10(t *testing.T) {
	cases := []struct {
		ord, want int
	}{
		{22, 10}, // AAA
		{21, 9},  // AA+
		{20, 8},  // AA
		{19, 7},  // AA-
		{18, 6},  // A+
		{17, 5},  // A
		{16, 5},  // A-
		{15, 4},  // BBB+
		{14, 4},  // BBB
		{13, 3},  // BBB-
		{12, 3},  // BB+
		{11, 2},  // BB
		{10, 2},  // BB-
		{9, 1},   // B+
		{8, 1},   // B
		{7, 1},   // B-
		{6, 0},   // CCC+
		{0, 0},   // unrated
	}
	for _, c := range cases {
		if got := LegacyScore10(c.ord); got != c.want {
			t.Errorf("LegacyScore10(%d) = %d, want %d", c.ord, got, c.want)
		}
	}
}

// ДОХОДЪ scores must round-trip through parseDohodNumeric → LegacyScore10
// so the legacy-score-based UI filters keep behaving as before.
func TestDohodLegacyRoundTrip(t *testing.T) {
	for n := 1; n <= 10; n++ {
		ord := parseDohodNumeric(strconv.Itoa(n))
		if got := LegacyScore10(ord); got != n {
			t.Errorf("ДОХОДЪ %d → ord %d → legacy %d, want %d", n, ord, got, n)
		}
	}
}
