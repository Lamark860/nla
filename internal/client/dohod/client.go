package dohod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"nla/internal/model"
)

var nuxtDataRe = regexp.MustCompile(`(?s)id="__NUXT_DATA__"[^>]*>(.*?)</script>`)

// Client fetches and parses bond data from analytics.dohod.ru
type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://analytics.dohod.ru",
	}
}

// FetchBond retrieves bond + emitter data by ISIN from dohod.ru
func (c *Client) FetchBond(ctx context.Context, isin string) (*model.DohodBondData, error) {
	if !isValidISIN(isin) {
		return nil, fmt.Errorf("invalid ISIN: %s", isin)
	}

	url := fmt.Sprintf("%s/bond/%s", c.baseURL, isin)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "NLA-BondAnalyzer/1.0")
	req.Header.Set("Accept", "text/html")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dohod request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dohod returned %d for %s", resp.StatusCode, isin)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return ParseNuxtPayload(body, isin)
}

// ParseNuxtPayload extracts bond data from Nuxt3 SSR __NUXT_DATA__ payload.
// Exported for testing.
func ParseNuxtPayload(html []byte, isin string) (*model.DohodBondData, error) {
	match := nuxtDataRe.FindSubmatch(html)
	if match == nil {
		return nil, fmt.Errorf("__NUXT_DATA__ not found in HTML")
	}

	var payload []any
	if err := json.Unmarshal(match[1], &payload); err != nil {
		return nil, fmt.Errorf("parse NUXT payload: %w", err)
	}

	// Find the main bond data object (has 'isin' and 'issuer' keys)
	for i, item := range payload {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if _, hasISIN := m["isin"]; !hasISIN {
			continue
		}
		if _, hasIssuer := m["issuer"]; !hasIssuer {
			continue
		}

		resolved := resolveNuxtRefs(payload, i, 0)
		rm, ok := resolved.(map[string]any)
		if !ok {
			continue
		}

		return mapToDohodData(rm, isin)
	}

	return nil, fmt.Errorf("bond data object not found in payload")
}

// resolveNuxtRefs recursively resolves Nuxt3 payload references.
// Nuxt3 stores data as a flat array with integer references between items.
func resolveNuxtRefs(payload []any, idx int, depth int) any {
	if depth > 12 || idx < 0 || idx >= len(payload) {
		return nil
	}

	item := payload[idx]

	switch v := item.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, val := range v {
			if ref, ok := toInt(val); ok {
				result[key] = resolveNuxtRefs(payload, ref, depth+1)
			} else {
				result[key] = val
			}
		}
		return result

	case []any:
		// Nuxt reactive wrapper: ["ShallowReactive", 42]
		if len(v) == 2 {
			if tag, ok := v[0].(string); ok {
				if tag == "ShallowReactive" || tag == "Reactive" || tag == "ShallowRef" || tag == "Ref" {
					if ref, ok := toInt(v[1]); ok {
						return resolveNuxtRefs(payload, ref, depth+1)
					}
				}
			}
		}
		return item

	default:
		return item
	}
}

func mapToDohodData(m map[string]any, isin string) (*model.DohodBondData, error) {
	now := time.Now()

	d := &model.DohodBondData{
		ISIN:      isin,
		FetchedAt: now,
		UpdatedAt: now,
	}

	// Issuer
	if issuer, ok := m["issuer"].(map[string]any); ok {
		d.IssuerName = getString(issuer, "nameShort")
		d.IssuerSector = getString(issuer, "economySector")
		d.Country = getString(issuer, "countryName")
	}
	d.BorrowerName = getString(m, "borrower")

	// Credit ratings
	d.CreditRating = getFloat(m, "creditRating")
	d.CreditRatingText = getString(m, "creditRatingText")
	d.AKRA = getString(m, "akra")
	d.ExpertRA = getString(m, "expert")
	d.Fitch = getString(m, "fitch")
	d.Moody = getString(m, "moody")
	d.SP = getString(m, "sp")

	// DOHOD ratings
	d.EstimationRating = getFloat(m, "estimationRating")
	d.EstimationRatingText = getString(m, "estimationRatingText")

	// Quality
	d.Quality = getFloatPtr(m, "quality")
	d.QualityOutside = getFloatPtr(m, "qout")
	d.QualityInside = getFloatPtr(m, "qin")
	d.QualityBalance = getFloatPtr(m, "qbalance")
	d.QualityEarnings = getFloatPtr(m, "qearnings")
	d.QualityROEScore = getFloatPtr(m, "qratingROE")
	d.QualityROEValue = getFloatPtr(m, "qvalueROE")
	d.QualityNetDebt = getFloatPtr(m, "qratingNetDebtEquity")
	d.QualityNetDebtVal = getFloatPtr(m, "qvalueNetDebtEquity")
	d.QualityProfitChg = getFloatPtr(m, "qcol")

	// DP
	d.DP1 = getFloatPtr(m, "qdp1")
	d.DP2 = getFloatPtr(m, "qearningsDP2")
	d.DP3 = getFloatPtr(m, "qbalanceDP3")

	// Profitability
	d.ProfitROS = getFloatPtr(m, "qros")
	d.ProfitROSValue = getFloatPtr(m, "qvalueROS")
	d.ProfitOper = getFloatPtr(m, "qoperProf")
	d.ProfitOperVal = getFloatPtr(m, "qvalueOperProf")

	// Turnover
	d.TurnoverInventory = getFloatPtr(m, "qinventorTurnov")
	d.TurnoverCurAsset = getFloatPtr(m, "qturnovOfCurAsset")
	d.TurnoverReceiv = getFloatPtr(m, "qreceivableTurnov")

	// Liquidity
	d.LiqBalance = getFloatPtr(m, "qbalanceLiq")
	d.LiqCurrent = getFloatPtr(m, "qcurrentLiq")
	d.LiqQuick = getFloatPtr(m, "qquiqLiq")
	d.LiqCash = getFloatPtr(m, "qratingCashRatio")
	d.LiqCurrentVal = getFloatPtr(m, "qvalueCurrentLiq")
	d.LiqQuickVal = getFloatPtr(m, "qvalueQuiqLiq")
	d.LiqCashVal = getFloatPtr(m, "qvalueCashRatio")

	// Stability
	d.Stability = getFloatPtr(m, "qstability")
	d.StabilityDebt = getFloatPtr(m, "qshortliabilities")
	d.StabDebtVal = getFloatPtr(m, "qvalueShortliabilities")

	// Key metrics
	d.BestScore = getFloatPtr(m, "bestScore")
	d.DownRisk = getFloatPtr(m, "downRisk")
	d.Liquidity = getFloatPtr(m, "liquidity")
	d.TotalReturn = getFloatPtr(m, "totalReturn")
	d.CurrentYield = getFloatPtr(m, "currentYield")
	d.Size = getFloatPtr(m, "size")
	d.Complexity = getFloatPtr(m, "complexity")

	// Bond description fields
	d.Description = getString(m, "description")
	d.Event = getString(m, "event")
	d.CouponRate = getFloatPtr(m, "couponRate")
	d.CouponRateAfterPut = getFloatPtr(m, "couponRateAfterPut")
	d.CouponSize = getFloatPtr(m, "couponSize")
	d.EarlyRedemptionCall = getString(m, "earlyRedemptionCallDate")
	d.YearsToMaturity = getFloatPtr(m, "yearsToMaturity")
	d.Duration = getFloatPtr(m, "duration")
	d.DurationMd = getFloatPtr(m, "durationMd")
	d.SimpleYield = getFloatPtr(m, "simpleYield")
	d.ForQualifiedOnly = getBool(m, "forQualifiedInvestors")
	d.TaxLongtermFree = getBool(m, "taxLongtermFree")
	d.TaxFree = getBool(m, "taxFree")
	d.TaxCurrencyFree = getBool(m, "taxCurrencyFree")
	d.SectorText = getString(m, "sectorText")
	d.MinLot = getFloatPtr(m, "minlot")
	d.FRNIndex = getString(m, "frnIndex")
	d.FRNIndexAdd = getFloatPtr(m, "frnIndexAdd")
	d.FRNFormulaText = getString(m, "frnFormulaText")

	return d, nil
}

// --- helpers ---

func isValidISIN(isin string) bool {
	if len(isin) != 12 {
		return false
	}
	for _, c := range isin {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func getFloatPtr(m map[string]any, key string) *float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return &f
		}
	}
	return nil
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case float64:
			return b != 0
		}
	}
	return false
}
