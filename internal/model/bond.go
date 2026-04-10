package model

import "time"

// Bond represents a bond with MOEX data and calculated fields
type Bond struct {
	// MOEX security fields
	SECID         string  `json:"secid"`
	ShortName     string  `json:"shortname"`
	SecName       string  `json:"secname"`
	ISIN          string  `json:"isin"`
	FaceValue     float64 `json:"facevalue"`
	MatDate       string  `json:"matdate"`
	CouponPeriod  int     `json:"coupon_period"`
	CouponValue   float64 `json:"coupon_value"`
	CouponPercent float64 `json:"coupon_percent"`
	NextCoupon    string  `json:"next_coupon"`
	AccruedInt    float64 `json:"accrued_int"`
	BondType      string  `json:"bond_type"`

	// Extended securities fields (from MOEX securities block)
	BoardID             string   `json:"boardid"`
	BoardName           string   `json:"boardname"`
	LatName             string   `json:"latname"`
	RegNumber           string   `json:"regnumber"`
	CurrencyID          string   `json:"currencyid"`
	FaceUnit            string   `json:"faceunit"`
	IssueSize           int64    `json:"issuesize"`
	IssueSizePlaced     int64    `json:"issuesize_placed"`
	LotSize             int      `json:"lotsize"`
	LotValue            float64  `json:"lotvalue"`
	MinStep             float64  `json:"minstep"`
	Decimals            int      `json:"decimals"`
	ListLevel           int      `json:"listlevel"`
	SecType             string   `json:"sectype"`
	BondTypeName        string   `json:"bondtype_name"`
	BondSubType         string   `json:"bondsubtype"`
	SectorID            string   `json:"sectorid"`
	MarketCode          string   `json:"marketcode"`
	InstrID             string   `json:"instrid"`
	SettleDate          string   `json:"settledate"`
	OfferDate           string   `json:"offerdate"`
	BuyBackDate         string   `json:"buybackdate"`
	BuyBackPrice        *float64 `json:"buybackprice"`
	CallOptionDate      string   `json:"calloptiondate"`
	PutOptionDate       string   `json:"putoptiondate"`
	PrevWAPrice         *float64 `json:"prevwaprice"`
	YieldAtPrevWAPrice  *float64 `json:"yieldatprevwaprice"`
	PrevLegalClosePrice *float64 `json:"prevlegalcloseprice"`
	PrevDate            string   `json:"prevdate"`
	FaceValueOnSettle   float64  `json:"facevalue_on_settle"`

	// MOEX marketdata fields (prices in % of face value)
	Last     *float64 `json:"last"`
	Bid      *float64 `json:"bid"`
	Offer    *float64 `json:"offer"`
	Yield    *float64 `json:"yield"`
	Duration *int     `json:"duration"`
	VolToday int64    `json:"vol_today"`

	// Extended marketdata fields for Trading tab
	Open       *float64 `json:"open"`
	Low        *float64 `json:"low"`
	High       *float64 `json:"high"`
	WAPrice    *float64 `json:"waprice"`
	NumTrades  *int64   `json:"numtrades"`
	ValToday   *float64 `json:"valtoday"`
	BidDepth   *int64   `json:"biddeptht"`
	OfferDepth *int64   `json:"offerdeptht"`
	NumBids    *int64   `json:"numbids"`
	NumOffers  *int64   `json:"numoffers"`
	UpdateTime string   `json:"updatetime"`
	TradeTime  string   `json:"tradetime"`
	SysTime    string   `json:"systime"`
	PrevPrice  *float64 `json:"prevprice"`

	// Extended marketdata yield fields
	Spread                *float64 `json:"spread"`
	YieldAtWAPrice        *float64 `json:"yieldatwaprice"`
	YieldToPrevYield      *float64 `json:"yieldtoprevyield"`
	CloseYield            *float64 `json:"closeyield"`
	EffectiveYield        *float64 `json:"effectiveyield"`
	LastToLastWAPrice     *float64 `json:"lastcngtolastwaprice"`
	WAPToPrevWAPricePrcnt *float64 `json:"waptoprevwapriceprcnt"`
	WAPToPrevWAPrice      *float64 `json:"waptoprevwaprice"`
	LastToPrevPrice       *float64 `json:"lasttoprevprice"`
	PriceMinusPrevWAPrice *float64 `json:"priceminusprevwaprice"`
	MarketPrice           *float64 `json:"marketprice"`
	MarketPriceToday      *float64 `json:"marketpricetoday"`
	LCurrentPrice         *float64 `json:"lcurrentprice"`
	LClosePrice           *float64 `json:"lcloseprice"`
	Change                *float64 `json:"change"`
	YieldToOffer          *float64 `json:"yieldtooffer"`
	YieldLastCoupon       *float64 `json:"yieldlastcoupon"`
	ZSpread               *float64 `json:"zspread"`
	ZSpreadAtWAPrice      *float64 `json:"zspreadatwaprice"`
	IRICPIClose           *float64 `json:"iricpiclose"`
	BEIClose              *float64 `json:"beiclose"`
	CBRClose              *float64 `json:"cbrclose"`
	CallOptionYield       *float64 `json:"calloptionyield"`
	CallOptionDuration    *int     `json:"calloptionduration"`
	DurationWAPrice       *int     `json:"durationwaprice"`

	// Price change fields
	LastChange      *float64 `json:"last_change"`
	LastChangePrcnt *float64 `json:"last_change_prcnt"`

	// Trading status
	TradingStatus string `json:"trading_status"` // T=trading, N=not trading, S=suspended

	// Calculated fields
	PriceRUB       *float64 `json:"price_rub"`
	ValueTodayRUB  *float64 `json:"value_today_rub"`
	DaysToMaturity *int     `json:"days_to_maturity"`
	IsFloat        bool     `json:"is_float"`
	IsIndexed      bool     `json:"is_indexed"`
	BondCategory   string   `json:"bond_category"`
	CouponDisplay  *float64 `json:"coupon_display"`

	// Extended calculated fields (ASH parity)
	YearsToMaturity  *float64 `json:"years_to_maturity"`
	DaysToCall       *int     `json:"days_to_call"`
	DaysToPut        *int     `json:"days_to_put"`
	DaysToNextCoupon *int     `json:"days_to_next_coupon"`
	CurrentYield     *float64 `json:"current_yield"`
	ModifiedDuration *float64 `json:"modified_duration"`
	SpreadAbsolute   *float64 `json:"spread_absolute"`
	SpreadPercent    *float64 `json:"spread_percent"`
	MidPricePct      *float64 `json:"mid_price_pct"`
	MidPriceRUB      *float64 `json:"mid_price_rub"`
	BidRUB           *float64 `json:"bid_rub"`
	OfferRUB         *float64 `json:"offer_rub"`
	AvgTradeSize     *float64 `json:"avg_trade_size"`
	TotalDepth       *int64   `json:"total_depth"`
	BidOfferRatio    *float64 `json:"bid_offer_ratio"`
	AccruedIntPct    *float64 `json:"accrued_int_pct"`
	LifeProgress     *float64 `json:"life_progress"`
	IsNearOffer      bool     `json:"is_near_offer"`
	HasNoFixedCoupon bool     `json:"has_no_fixed_coupon"`
	TradingStatusTxt string   `json:"trading_status_text"`
	RiskCategory     string   `json:"risk_category"`

	// Issuer info (resolved from bond_issuers collection)
	EmitterID   *int64 `json:"emitter_id,omitempty"`
	EmitterName string `json:"emitter_name,omitempty"`
}

// BondDetail represents full bond details from MOEX with coupons and history
type BondDetail struct {
	Bond        Bond     `json:"bond"`
	Coupons     []Coupon `json:"coupons,omitempty"`
	History     []OHLC   `json:"history,omitempty"`
	EmitterID   *int64   `json:"emitter_id,omitempty"`
	EmitterName string   `json:"emitter_name,omitempty"`
}

// Coupon represents a single coupon payment
type Coupon struct {
	CouponDate   string  `json:"coupon_date"`
	RecordDate   string  `json:"record_date"`
	StartDate    string  `json:"start_date"`
	Value        float64 `json:"value"`
	ValuePercent float64 `json:"value_percent"`
	ValueRUB     float64 `json:"value_rub"`
}

// OHLC represents a candlestick for bond price history
type OHLC struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int64   `json:"volume"`
	Value  float64 `json:"value"`
}

// BondListResponse is the paginated bonds list
type BondListResponse struct {
	Data []Bond  `json:"data"`
	Meta PagMeta `json:"meta"`
}

type PagMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// BondAnalysis represents an AI analysis stored in MongoDB
type BondAnalysis struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	SECID      string    `json:"secid" bson:"secid"`
	Response   string    `json:"response" bson:"response"`
	Rating     *float64  `json:"rating" bson:"rating"`
	JSONData   any       `json:"json_data,omitempty" bson:"json_data,omitempty"`
	CustomJSON any       `json:"custom_json,omitempty" bson:"custom_json,omitempty"`
	Timestamp  time.Time `json:"timestamp" bson:"timestamp"`
	UserID     *int64    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	SavedAt    time.Time `json:"saved_at" bson:"saved_at"`
	Tags       []string  `json:"tags,omitempty" bson:"tags,omitempty"`
}

// AnalysisStats — aggregate stats for a bond's analyses
type AnalysisStats struct {
	Total        int        `json:"total"`
	AvgRating    float64    `json:"avg_rating"`
	LastAnalysis *time.Time `json:"last_analysis"`
}

// BondIssuer maps SECID to emitter
type BondIssuer struct {
	SECID       string    `json:"secid" bson:"secid"`
	EmitterID   int64     `json:"emitter_id" bson:"emitter_id"`
	EmitterName string    `json:"emitter_name" bson:"emitter_name"`
	IsHidden    bool      `json:"is_hidden" bson:"is_hidden"`
	NeedsSync   bool      `json:"needs_sync" bson:"needs_sync"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// IssuerRating holds credit rating for an issuer
type IssuerRating struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	EmitterID int64     `json:"emitter_id" bson:"emitter_id"`
	Issuer    string    `json:"issuer" bson:"issuer"`
	Agency    string    `json:"agency" bson:"agency"`
	Rating    string    `json:"rating" bson:"rating"`
	Score     int       `json:"score" bson:"score"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// IssuerRatingResponse is the API response for issuer ratings
type IssuerRatingResponse struct {
	EmitterID int64          `json:"emitter_id"`
	Issuer    string         `json:"issuer"`
	Ratings   []IssuerRating `json:"ratings"`
	Score     int            `json:"score"`
}

// IssuerGroup represents a group of bonds by the same emitter
type IssuerGroup struct {
	EmitterID   int64  `json:"emitter_id"`
	EmitterName string `json:"emitter_name"`
	BondCount   int    `json:"bond_count"`
	Bonds       []Bond `json:"bonds"`
}

// IssuerGroupResponse is the paginated grouped bonds response
type IssuerGroupResponse struct {
	Groups       []IssuerGroup `json:"groups"`
	TotalIssuers int           `json:"total_issuers"`
	TotalBonds   int           `json:"total_bonds"`
}
