package model

import "time"

// DohodBondData contains parsed data from analytics.dohod.ru for a bond
type DohodBondData struct {
	ISIN  string `json:"isin" bson:"isin"`
	Secid string `json:"secid" bson:"secid"`

	// Issuer info
	IssuerName   string `json:"issuer_name" bson:"issuer_name"`
	IssuerSector string `json:"issuer_sector" bson:"issuer_sector"`
	BorrowerName string `json:"borrower_name" bson:"borrower_name"`
	Country      string `json:"country" bson:"country"`

	// Credit ratings (agency → rating string)
	CreditRating     float64 `json:"credit_rating" bson:"credit_rating"`           // score 1-10
	CreditRatingText string  `json:"credit_rating_text" bson:"credit_rating_text"` // e.g. "BBB+"
	AKRA             string  `json:"akra" bson:"akra"`
	ExpertRA         string  `json:"expert_ra" bson:"expert_ra"`
	Fitch            string  `json:"fitch" bson:"fitch"`
	Moody            string  `json:"moody" bson:"moody"`
	SP               string  `json:"sp" bson:"sp"`

	// DOHOD.ru own ratings
	EstimationRating     float64 `json:"estimation_rating" bson:"estimation_rating"`
	EstimationRatingText string  `json:"estimation_rating_text" bson:"estimation_rating_text"`

	// Quality scores
	Quality           *float64 `json:"quality" bson:"quality"`
	QualityOutside    *float64 `json:"quality_outside" bson:"quality_outside"`
	QualityInside     *float64 `json:"quality_inside" bson:"quality_inside"`
	QualityBalance    *float64 `json:"quality_balance" bson:"quality_balance"`
	QualityEarnings   *float64 `json:"quality_earnings" bson:"quality_earnings"`
	QualityROEScore   *float64 `json:"quality_roe_score" bson:"quality_roe_score"`
	QualityROEValue   *float64 `json:"quality_roe_value" bson:"quality_roe_value"`
	QualityNetDebt    *float64 `json:"quality_net_debt" bson:"quality_net_debt"`
	QualityNetDebtVal *float64 `json:"quality_net_debt_value" bson:"quality_net_debt_value"`
	QualityProfitChg  *float64 `json:"quality_profit_change" bson:"quality_profit_change"`

	// Penalties/bonuses (DP1, DP2, DP3)
	DP1 *float64 `json:"dp1" bson:"dp1"`
	DP2 *float64 `json:"dp2" bson:"dp2"`
	DP3 *float64 `json:"dp3" bson:"dp3"`

	// Profitability
	ProfitROS      *float64 `json:"profit_ros" bson:"profit_ros"`
	ProfitROSValue *float64 `json:"profit_ros_value" bson:"profit_ros_value"`
	ProfitOper     *float64 `json:"profit_oper" bson:"profit_oper"`
	ProfitOperVal  *float64 `json:"profit_oper_value" bson:"profit_oper_value"`

	// Turnover
	TurnoverInventory *float64 `json:"turnover_inventory" bson:"turnover_inventory"`
	TurnoverCurAsset  *float64 `json:"turnover_cur_asset" bson:"turnover_cur_asset"`
	TurnoverReceiv    *float64 `json:"turnover_receivable" bson:"turnover_receivable"`

	// Liquidity
	LiqBalance *float64 `json:"liq_balance" bson:"liq_balance"`
	LiqCurrent *float64 `json:"liq_current" bson:"liq_current"`
	LiqQuick   *float64 `json:"liq_quick" bson:"liq_quick"`
	LiqCash    *float64 `json:"liq_cash_ratio" bson:"liq_cash_ratio"`

	// Liquidity values
	LiqCurrentVal *float64 `json:"liq_current_value" bson:"liq_current_value"`
	LiqQuickVal   *float64 `json:"liq_quick_value" bson:"liq_quick_value"`
	LiqCashVal    *float64 `json:"liq_cash_value" bson:"liq_cash_value"`

	// Stability
	Stability     *float64 `json:"stability" bson:"stability"`
	StabilityDebt *float64 `json:"stability_short_debt" bson:"stability_short_debt"`
	StabDebtVal   *float64 `json:"stability_debt_value" bson:"stability_debt_value"`

	// Key metrics
	BestScore    *float64 `json:"best_score" bson:"best_score"`
	DownRisk     *float64 `json:"down_risk" bson:"down_risk"`
	Liquidity    *float64 `json:"liquidity_score" bson:"liquidity_score"`
	TotalReturn  *float64 `json:"total_return" bson:"total_return"`
	CurrentYield *float64 `json:"current_yield" bson:"current_yield"`
	Size         *float64 `json:"size" bson:"size"`
	Complexity   *float64 `json:"complexity" bson:"complexity"`

	// Bond description from dohod.ru
	Description         string   `json:"description,omitempty" bson:"description,omitempty"`
	Event               string   `json:"event,omitempty" bson:"event,omitempty"`                                 // ближайшее событие (погашение/put/call)
	CouponRate          *float64 `json:"coupon_rate,omitempty" bson:"coupon_rate,omitempty"`                     // ставка купона
	CouponRateAfterPut  *float64 `json:"coupon_rate_after_put,omitempty" bson:"coupon_rate_after_put,omitempty"` // ставка после оферты
	CouponSize          *float64 `json:"coupon_size,omitempty" bson:"coupon_size,omitempty"`
	EarlyRedemptionCall string   `json:"early_redemption_call,omitempty" bson:"early_redemption_call,omitempty"`
	YearsToMaturity     *float64 `json:"years_to_maturity,omitempty" bson:"years_to_maturity,omitempty"`
	Duration            *float64 `json:"dohod_duration,omitempty" bson:"dohod_duration,omitempty"`
	DurationMd          *float64 `json:"dohod_duration_md,omitempty" bson:"dohod_duration_md,omitempty"`
	SimpleYield         *float64 `json:"simple_yield,omitempty" bson:"simple_yield,omitempty"`
	ForQualifiedOnly    bool     `json:"for_qualified_only" bson:"for_qualified_only"`
	TaxLongtermFree     bool     `json:"tax_longterm_free" bson:"tax_longterm_free"`
	TaxFree             bool     `json:"tax_free" bson:"tax_free"`
	TaxCurrencyFree     bool     `json:"tax_currency_free" bson:"tax_currency_free"`
	SectorText          string   `json:"sector_text,omitempty" bson:"sector_text,omitempty"`
	MinLot              *float64 `json:"min_lot,omitempty" bson:"min_lot,omitempty"`
	FRNIndex            string   `json:"frn_index,omitempty" bson:"frn_index,omitempty"`         // для флоатеров: базовая ставка
	FRNIndexAdd         *float64 `json:"frn_index_add,omitempty" bson:"frn_index_add,omitempty"` // надбавка к ставке
	FRNFormulaText      string   `json:"frn_formula_text,omitempty" bson:"frn_formula_text,omitempty"`

	// Timestamps
	FetchedAt time.Time `json:"fetched_at" bson:"fetched_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
