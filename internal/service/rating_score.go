package service

import (
	"regexp"
	"strconv"
	"strings"
)

// Ordinal credit-rating scale (higher = better). All agency notations
// (АКРА, Эксперт РА, НКР, НРА, S&P, Fitch, Moody's, ДОХОДЪ) collapse onto
// this single scale so ratings are comparable and sortable across agencies.
//
//	AAA=22
//	AA+=21, AA=20, AA-=19
//	A+=18,  A=17,  A-=16
//	BBB+=15, BBB=14, BBB-=13
//	BB+=12,  BB=11,  BB-=10
//	B+=9,    B=8,    B-=7
//	CCC+=6,  CCC=5,  CCC-=4
//	CC=3
//	C=2
//	D=1 (RD also folds here)
//	0 = unrated / unparseable / withdrawn
const (
	OrdMax = 22
	OrdMin = 0
)

var ordToTier = map[int]string{
	22: "AAA",
	21: "AA", 20: "AA", 19: "AA",
	18: "A", 17: "A", 16: "A",
	15: "BBB", 14: "BBB", 13: "BBB",
	12: "BB", 11: "BB", 10: "BB",
	9: "B", 8: "B", 7: "B",
	6: "CCC", 5: "CCC", 4: "CCC",
	3: "CC",
	2: "C",
	1: "D",
}

var (
	moodysRe = regexp.MustCompile(`^(Aaa|Aa[1-3]|A[1-3]|Baa[1-3]|Ba[1-3]|B[1-3]|Caa[1-3]|Ca|C)$`)
	letterRe = regexp.MustCompile(`^(AAA|AA[+-]?|A[+-]?|BBB[+-]?|BB[+-]?|B[+-]?|CCC[+-]?|CC|C|D|RD)$`)
)

// NormalizeRating parses a credit rating string from any of the agencies we
// encounter and returns its ordinal score on the 1-22 scale and a canonical
// tier letter group ("AAA", "AA", ..., "D"). Returns (0, "") if the input
// cannot be classified.
//
// Accepted formats (case as appropriate per agency):
//
//	АКРА        AAA(RU), AA+(RU)
//	Эксперт РА  ruAAA, ruAA+
//	НКР         AAA.ru, AA+.ru
//	НРА         AAA|ru|, AA+|ru|
//	S&P/Fitch   AAA, AA+, ..., D
//	Moody's     Aaa, Aa1, A2, Baa3, Ba1, B2, Caa3, Ca, C
//	ДОХОДЪ      "8" or "8/10" (numeric 1-10)
//
// Outlook words ("Stable", "Positive", "Негативный", ...) and parenthetical
// suffixes are stripped before classification.
func NormalizeRating(text string) (ord int, tier string) {
	s := strings.TrimSpace(text)
	if s == "" {
		return 0, ""
	}

	low := strings.ToLower(s)
	if low == "отозван" || low == "отозвано" || strings.HasPrefix(low, "wd") || s == "—" {
		return 0, ""
	}

	if n := parseDohodNumeric(s); n > 0 {
		return n, ordToTier[n]
	}

	s = stripOutlook(s)
	s = stripNationalSuffix(s)

	// "ru"/"Ru"/"RU" prefix used by Эксперт РА and some MOEX CCI variants
	if len(s) > 2 && strings.EqualFold(s[:2], "ru") {
		s = s[2:]
	}
	s = strings.TrimSpace(s)

	// Try Moody's (case-sensitive: only "Aaa" / "Baa1" patterns) first
	if ord := parseMoodys(s); ord > 0 {
		return ord, ordToTier[ord]
	}

	if ord := parseLetterScale(strings.ToUpper(s)); ord > 0 {
		return ord, ordToTier[ord]
	}

	return 0, ""
}

// LegacyScore10 maps the 22-level ordinal back to the 1-10 scale that the
// existing API and frontend filters depend on.
//
// Mapping is chosen so that round-tripping a ДОХОДЪ score "n" (1..10) through
// parseDohodNumeric → LegacyScore10 returns "n".
func LegacyScore10(ord int) int {
	switch {
	case ord >= 22:
		return 10
	case ord == 21:
		return 9
	case ord == 20:
		return 8
	case ord == 19:
		return 7
	case ord == 18:
		return 6
	case ord == 17, ord == 16:
		return 5
	case ord == 15, ord == 14:
		return 4
	case ord == 13, ord == 12:
		return 3
	case ord == 11, ord == 10:
		return 2
	case ord >= 7:
		return 1
	default:
		return 0
	}
}

// parseDohodNumeric parses ДОХОДЪ "X" or "X/10" into the 22-level ordinal.
// Mapping chosen so LegacyScore10(parseDohodNumeric(n)) == n for n in 1..10.
func parseDohodNumeric(s string) int {
	clean := strings.TrimSpace(s)
	if i := strings.Index(clean, "/"); i > 0 {
		clean = strings.TrimSpace(clean[:i])
	}
	n, err := strconv.Atoi(clean)
	if err != nil || n < 1 || n > 10 {
		return 0
	}
	return map[int]int{10: 22, 9: 21, 8: 20, 7: 19, 6: 18, 5: 17, 4: 15, 3: 13, 2: 11, 1: 9}[n]
}

func parseMoodys(s string) int {
	if !moodysRe.MatchString(s) {
		return 0
	}
	return map[string]int{
		"Aaa": 22,
		"Aa1": 21, "Aa2": 20, "Aa3": 19,
		"A1": 18, "A2": 17, "A3": 16,
		"Baa1": 15, "Baa2": 14, "Baa3": 13,
		"Ba1": 12, "Ba2": 11, "Ba3": 10,
		"B1": 9, "B2": 8, "B3": 7,
		"Caa1": 6, "Caa2": 5, "Caa3": 4,
		"Ca": 3, "C": 2,
	}[s]
}

func parseLetterScale(s string) int {
	if !letterRe.MatchString(s) {
		return 0
	}
	return map[string]int{
		"AAA": 22,
		"AA+": 21, "AA": 20, "AA-": 19,
		"A+": 18, "A": 17, "A-": 16,
		"BBB+": 15, "BBB": 14, "BBB-": 13,
		"BB+": 12, "BB": 11, "BB-": 10,
		"B+": 9, "B": 8, "B-": 7,
		"CCC+": 6, "CCC": 5, "CCC-": 4,
		"CC": 3, "C": 2,
		"D": 1, "RD": 1,
	}[s]
}

// stripNationalSuffix removes "(RU)"/"(EXP)"/.. parens, ".ru"/".RU" (НКР),
// "|ru|"/"|RU|" (НРА). Anything inside or after a "(" is also dropped, which
// also catches outlook annotations like "AAA (Negative)".
func stripNationalSuffix(s string) string {
	if i := strings.Index(s, "("); i > 0 {
		s = s[:i]
	}
	for _, suf := range []string{".ru", ".RU", "|ru|", "|RU|"} {
		s = strings.TrimSuffix(s, suf)
	}
	return strings.TrimSpace(s)
}

var outlookWords = []string{
	"стабильный", "позитивный", "негативный", "развивающийся",
	"стаб.", "позит.", "негат.", "разв.",
	"stable", "positive", "negative", "developing",
}

func stripOutlook(s string) string {
	low := strings.ToLower(s)
	for _, w := range outlookWords {
		if i := strings.Index(low, w); i > 0 {
			return strings.TrimSpace(s[:i])
		}
	}
	return s
}
