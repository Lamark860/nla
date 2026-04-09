package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"nla/internal/client/moex"
	"nla/internal/model"
	mongorepo "nla/internal/mongo"
)

const bondsCacheKey = "bonds:list"
const bondCachePrefix = "bonds:"
const bondsCacheTTL = 24 * time.Hour

type BondService struct {
	moex       *moex.Client
	redis      *redis.Client
	issuerRepo *mongorepo.IssuerRepo
}

func NewBondService(moexClient *moex.Client, rdb *redis.Client, issuerRepo *mongorepo.IssuerRepo) *BondService {
	return &BondService{moex: moexClient, redis: rdb, issuerRepo: issuerRepo}
}

// GetBondsPaginated returns sorted and paginated bonds
func (s *BondService) GetBondsPaginated(ctx context.Context, page, perPage int, sortBy string) (*model.BondListResponse, error) {
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return nil, err
	}

	sortBonds(bonds, sortBy)

	total := len(bonds)
	start := (page - 1) * perPage
	if start >= total {
		return &model.BondListResponse{
			Data: []model.Bond{},
			Meta: model.PagMeta{Page: page, PerPage: perPage, Total: total},
		}, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return &model.BondListResponse{
		Data: bonds[start:end],
		Meta: model.PagMeta{Page: page, PerPage: perPage, Total: total},
	}, nil
}

// GetBondDetail returns full details for a single bond
func (s *BondService) GetBondDetail(ctx context.Context, secid string) (*model.Bond, error) {
	cacheKey := bondCachePrefix + secid
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var bond model.Bond
		if json.Unmarshal([]byte(cached), &bond) == nil {
			return &bond, nil
		}
	}

	raw, err := s.moex.GetSecurity(ctx, secid)
	if err != nil {
		return nil, fmt.Errorf("fetch bond %s: %w", secid, err)
	}

	securities := extractRows(raw, "securities")
	marketdata := extractRows(raw, "marketdata")

	if len(securities) == 0 {
		return nil, fmt.Errorf("bond %s not found", secid)
	}

	bond := parseBond(securities[0], safeFirst(marketdata))
	calculateFields(&bond)

	if data, err := json.Marshal(bond); err == nil {
		s.redis.Set(ctx, cacheKey, data, bondsCacheTTL)
	}

	return &bond, nil
}

// GetBondCoupons returns coupon schedule for a bond
func (s *BondService) GetBondCoupons(ctx context.Context, secid string) ([]model.Coupon, error) {
	raw, err := s.moex.GetBondization(ctx, secid)
	if err != nil {
		return nil, fmt.Errorf("fetch coupons %s: %w", secid, err)
	}

	rows := extractRows(raw, "coupons")
	coupons := make([]model.Coupon, 0, len(rows))
	for _, row := range rows {
		coupons = append(coupons, model.Coupon{
			CouponDate:   getString(row, "coupondate"),
			RecordDate:   getString(row, "recorddate"),
			StartDate:    getString(row, "startdate"),
			Value:        getFloat(row, "value"),
			ValuePercent: getFloat(row, "valueprc"),
			ValueRUB:     getFloat(row, "value_rub"),
		})
	}

	return coupons, nil
}

// GetBondHistory returns OHLC price history (180 days)
func (s *BondService) GetBondHistory(ctx context.Context, secid string) ([]model.OHLC, error) {
	from := time.Now().AddDate(0, -6, 0).Format("2006-01-02")
	till := time.Now().Format("2006-01-02")

	raw, err := s.moex.GetCandles(ctx, secid, from, till)
	if err != nil {
		return nil, fmt.Errorf("fetch history %s: %w", secid, err)
	}

	rows := extractRows(raw, "candles")
	history := make([]model.OHLC, 0, len(rows))
	for _, row := range rows {
		history = append(history, model.OHLC{
			Date:   getString(row, "begin"),
			Open:   getFloat(row, "open"),
			Close:  getFloat(row, "close"),
			High:   getFloat(row, "high"),
			Low:    getFloat(row, "low"),
			Volume: getInt64(row, "volume"),
			Value:  getFloat(row, "value"),
		})
	}

	return history, nil
}

// GetMonthlyBonds returns bonds with monthly coupon (period 27-33 days)
func (s *BondService) GetMonthlyBonds(ctx context.Context) ([]model.Bond, error) {
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return nil, err
	}

	monthly := make([]model.Bond, 0)
	for _, b := range bonds {
		if b.CouponPeriod >= 27 && b.CouponPeriod <= 33 {
			monthly = append(monthly, b)
		}
	}
	return monthly, nil
}

// GetBondsGroupedByIssuer returns bonds grouped by emitter_id from MongoDB bond_issuers
func (s *BondService) GetBondsGroupedByIssuer(ctx context.Context, monthlyOnly bool) (*model.IssuerGroupResponse, error) {
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return nil, err
	}

	if monthlyOnly {
		filtered := make([]model.Bond, 0)
		for _, b := range bonds {
			if b.CouponPeriod >= 27 && b.CouponPeriod <= 33 {
				filtered = append(filtered, b)
			}
		}
		bonds = filtered
	}

	// Get all secids
	secids := make([]string, len(bonds))
	for i, b := range bonds {
		secids[i] = b.SECID
	}

	// Batch fetch issuer mappings from MongoDB
	issuerMap, err := s.issuerRepo.GetBySecids(ctx, secids)
	if err != nil {
		return nil, fmt.Errorf("fetch issuer mappings: %w", err)
	}

	// Group by emitter_id
	type groupData struct {
		emitterID   int64
		emitterName string
		bonds       []model.Bond
	}
	groups := make(map[int64]*groupData)
	var noIssuerBonds []model.Bond

	for _, b := range bonds {
		if issuer, ok := issuerMap[b.SECID]; ok && issuer.EmitterID > 0 {
			g, exists := groups[issuer.EmitterID]
			if !exists {
				g = &groupData{
					emitterID:   issuer.EmitterID,
					emitterName: issuer.EmitterName,
					bonds:       make([]model.Bond, 0),
				}
				groups[issuer.EmitterID] = g
			}
			// Use non-empty emitter name
			if g.emitterName == "" && issuer.EmitterName != "" {
				g.emitterName = issuer.EmitterName
			}
			g.bonds = append(g.bonds, b)
		} else {
			noIssuerBonds = append(noIssuerBonds, b)
		}
	}

	// Convert to response
	result := make([]model.IssuerGroup, 0, len(groups))
	for _, g := range groups {
		name := g.emitterName
		if name == "" {
			name = fmt.Sprintf("Эмитент #%d", g.emitterID)
		}
		result = append(result, model.IssuerGroup{
			EmitterID:   g.emitterID,
			EmitterName: name,
			BondCount:   len(g.bonds),
			Bonds:       g.bonds,
		})
	}

	// Add ungrouped bonds as a special group
	if len(noIssuerBonds) > 0 {
		result = append(result, model.IssuerGroup{
			EmitterID:   0,
			EmitterName: "Без эмитента",
			BondCount:   len(noIssuerBonds),
			Bonds:       noIssuerBonds,
		})
	}

	// Sort by bond count desc (highest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].BondCount > result[j].BondCount
	})

	totalBonds := 0
	for _, g := range result {
		totalBonds += g.BondCount
	}

	return &model.IssuerGroupResponse{
		Groups:       result,
		TotalIssuers: len(result),
		TotalBonds:   totalBonds,
	}, nil
}

// getAllBonds loads all bonds from cache or MOEX
func (s *BondService) getAllBonds(ctx context.Context) ([]model.Bond, error) {
	if cached, err := s.redis.Get(ctx, bondsCacheKey).Result(); err == nil {
		var bonds []model.Bond
		if json.Unmarshal([]byte(cached), &bonds) == nil {
			return bonds, nil
		}
	}

	raw, err := s.moex.GetSecurities(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch securities: %w", err)
	}

	securities := extractRows(raw, "securities")
	marketdata := extractRows(raw, "marketdata")

	mdMap := make(map[string]map[string]any)
	for _, md := range marketdata {
		if secid := getString(md, "SECID"); secid != "" {
			mdMap[secid] = md
		}
	}

	bonds := make([]model.Bond, 0, len(securities))
	seen := make(map[string]bool)
	for _, sec := range securities {
		secid := getString(sec, "SECID")
		if seen[secid] {
			continue
		}
		seen[secid] = true
		md := mdMap[secid]
		bond := parseBond(sec, md)
		calculateFields(&bond)
		bonds = append(bonds, bond)
	}

	if data, err := json.Marshal(bonds); err == nil {
		s.redis.Set(ctx, bondsCacheKey, data, bondsCacheTTL)
	}

	return bonds, nil
}

// parseBond extracts bond fields from MOEX row maps
func parseBond(sec map[string]any, md map[string]any) model.Bond {
	b := model.Bond{
		SECID:         getString(sec, "SECID"),
		ShortName:     getString(sec, "SHORTNAME"),
		SecName:       getString(sec, "SECNAME"),
		ISIN:          getString(sec, "ISIN"),
		FaceValue:     getFloat(sec, "FACEVALUE"),
		MatDate:       getString(sec, "MATDATE"),
		CouponPeriod:  getInt(sec, "COUPONPERIOD"),
		CouponValue:   getFloat(sec, "COUPONVALUE"),
		CouponPercent: getFloat(sec, "COUPONPERCENT"),
		NextCoupon:    getString(sec, "NEXTCOUPON"),
		AccruedInt:    getFloat(sec, "ACCRUEDINT"),
		BondType:      getString(sec, "SECTYPE"),
		// Extended securities fields
		BoardID:             getString(sec, "BOARDID"),
		BoardName:           getString(sec, "BOARDNAME"),
		LatName:             getString(sec, "LATNAME"),
		RegNumber:           getString(sec, "REGNUMBER"),
		CurrencyID:          getString(sec, "CURRENCYID"),
		FaceUnit:            getString(sec, "FACEUNIT"),
		IssueSize:           getInt64(sec, "ISSUESIZE"),
		IssueSizePlaced:     getInt64(sec, "ISSUESIZEPLACED"),
		LotSize:             getInt(sec, "LOTSIZE"),
		LotValue:            getFloat(sec, "LOTVALUE"),
		MinStep:             getFloat(sec, "MINSTEP"),
		Decimals:            getInt(sec, "DECIMALS"),
		ListLevel:           getInt(sec, "LISTLEVEL"),
		SecType:             getString(sec, "SECTYPE"),
		BondTypeName:        getString(sec, "BONDTYPE"),
		BondSubType:         getString(sec, "BONDSUBTYPE"),
		SectorID:            getString(sec, "SECTORID"),
		MarketCode:          getString(sec, "MARKETCODE"),
		InstrID:             getString(sec, "INSTRID"),
		SettleDate:          getString(sec, "SETTLEDATE"),
		OfferDate:           getString(sec, "OFFERDATE"),
		BuyBackDate:         getString(sec, "BUYBACKDATE"),
		BuyBackPrice:        getFloatPtr(sec, "BUYBACKPRICE"),
		CallOptionDate:      getString(sec, "CALLOPTIONDATE"),
		PutOptionDate:       getString(sec, "PUTOPTIONDATE"),
		PrevWAPrice:         getFloatPtr(sec, "PREVWAPRICE"),
		YieldAtPrevWAPrice:  getFloatPtr(sec, "YIELDATPREVWAPRICE"),
		PrevLegalClosePrice: getFloatPtr(sec, "PREVLEGALCLOSEPRICE"),
		PrevDate:            getString(sec, "PREVDATE"),
		FaceValueOnSettle:   getFloat(sec, "FACEVALUEONSETTLEDATE"),
	}

	if md != nil {
		b.Last = getFloatPtr(md, "LAST")
		b.Bid = getFloatPtr(md, "BID")
		b.Offer = getFloatPtr(md, "OFFER")
		b.Yield = getFloatPtr(md, "YIELD")
		b.Duration = getIntPtr(md, "DURATION")
		b.VolToday = getInt64(md, "VOLTODAY")
		b.Open = getFloatPtr(md, "OPEN")
		b.Low = getFloatPtr(md, "LOW")
		b.High = getFloatPtr(md, "HIGH")
		b.WAPrice = getFloatPtr(md, "WAPRICE")
		b.NumTrades = getInt64Ptr(md, "NUMTRADES")
		b.ValToday = getFloatPtr(md, "VALTODAY")
		b.BidDepth = getInt64Ptr(md, "BIDDEPTHT")
		b.OfferDepth = getInt64Ptr(md, "OFFERDEPTHT")
		b.NumBids = getInt64Ptr(md, "NUMBIDS")
		b.NumOffers = getInt64Ptr(md, "NUMOFFERS")
		b.UpdateTime = getString(md, "UPDATETIME")
		b.TradeTime = getString(md, "TIME")
		b.SysTime = getString(md, "SYSTIME")
		b.PrevPrice = getFloatPtr(md, "PREVPRICE")
		b.LastChange = getFloatPtr(md, "LASTCHANGE")
		b.LastChangePrcnt = getFloatPtr(md, "LASTCHANGEPRCNT")
		b.TradingStatus = getString(md, "TRADINGSTATUS")
		// Extended marketdata yield/spread fields
		b.Spread = getFloatPtr(md, "SPREAD")
		b.YieldAtWAPrice = getFloatPtr(md, "YIELDATWAPRICE")
		b.YieldToPrevYield = getFloatPtr(md, "YIELDTOPREVYIELD")
		b.CloseYield = getFloatPtr(md, "CLOSEYIELD")
		b.LastToLastWAPrice = getFloatPtr(md, "LASTCNGTOLASTWAPRICE")
		b.WAPToPrevWAPricePrcnt = getFloatPtr(md, "WAPTOPREVWAPRICEPRCNT")
		b.WAPToPrevWAPrice = getFloatPtr(md, "WAPTOPREVWAPRICE")
		b.LastToPrevPrice = getFloatPtr(md, "LASTTOPREVPRICE")
		b.PriceMinusPrevWAPrice = getFloatPtr(md, "PRICEMINUSPREVWAPRICE")
		b.MarketPrice = getFloatPtr(md, "MARKETPRICE")
		b.MarketPriceToday = getFloatPtr(md, "MARKETPRICETODAY")
		b.LCurrentPrice = getFloatPtr(md, "LCURRENTPRICE")
		b.LClosePrice = getFloatPtr(md, "LCLOSEPRICE")
		b.Change = getFloatPtr(md, "CHANGE")
		b.YieldToOffer = getFloatPtr(md, "YIELDTOOFFER")
		b.YieldLastCoupon = getFloatPtr(md, "YIELDLASTCOUPON")
		b.ZSpread = getFloatPtr(md, "ZSPREAD")
		b.ZSpreadAtWAPrice = getFloatPtr(md, "ZSPREADATWAPRICE")
		b.IRICPIClose = getFloatPtr(md, "IRICPICLOSE")
		b.BEIClose = getFloatPtr(md, "BEICLOSE")
		b.CBRClose = getFloatPtr(md, "CBRCLOSE")
		b.CallOptionYield = getFloatPtr(md, "CALLOPTIONYIELD")
		b.CallOptionDuration = getIntPtr(md, "CALLOPTIONDURATION")
	}

	return b
}

// calculateFields computes derived values
func calculateFields(b *model.Bond) {
	// Price in RUB
	if b.Last != nil && b.FaceValue > 0 {
		rub := (*b.Last / 100) * b.FaceValue
		b.PriceRUB = &rub
	}

	// Value today in RUB
	if b.PriceRUB != nil && b.VolToday > 0 {
		val := float64(b.VolToday) * *b.PriceRUB
		b.ValueTodayRUB = &val
	}

	// Days to maturity
	if b.MatDate != "" {
		if matDate, err := time.Parse("2006-01-02", b.MatDate); err == nil {
			days := int(time.Until(matDate).Hours() / 24)
			b.DaysToMaturity = &days
		}
	}

	// Bond type flags
	bt := strings.ToLower(b.BondType)
	b.IsFloat = strings.Contains(bt, "флоатер") || strings.Contains(bt, "float")
	b.IsIndexed = strings.Contains(bt, "индексируемые") || b.BondType == "6"

	// Category
	switch {
	case strings.Contains(bt, "офз"):
		b.BondCategory = "ОФЗ"
	case strings.Contains(bt, "субфед"):
		b.BondCategory = "Субфедеральная"
	case strings.Contains(bt, "муницип"):
		b.BondCategory = "Муниципальная"
	default:
		b.BondCategory = "Корпоративная"
	}

	// Coupon display
	if b.CouponPercent > 0 {
		cp := b.CouponPercent
		b.CouponDisplay = &cp
	} else if b.CouponValue > 0 && b.FaceValue > 0 {
		cd := (b.CouponValue / b.FaceValue) * 100
		b.CouponDisplay = &cd
	}
}

// sortBonds sorts in-place by the given strategy
func sortBonds(bonds []model.Bond, sortBy string) {
	switch sortBy {
	case "yield_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].Yield) > safeFloat(bonds[j].Yield)
		})
	case "yield_asc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].Yield) < safeFloat(bonds[j].Yield)
		})
	case "maturity_asc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeInt(bonds[i].DaysToMaturity) < safeInt(bonds[j].DaysToMaturity)
		})
	case "maturity_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeInt(bonds[i].DaysToMaturity) > safeInt(bonds[j].DaysToMaturity)
		})
	case "volume_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return bonds[i].VolToday > bonds[j].VolToday
		})
	case "coupon_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].CouponDisplay) > safeFloat(bonds[j].CouponDisplay)
		})
	case "coupon_asc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].CouponDisplay) < safeFloat(bonds[j].CouponDisplay)
		})
	default: // "best" — composite
		sort.Slice(bonds, func(i, j int) bool {
			return bondScore(bonds[i]) > bondScore(bonds[j])
		})
	}
}

// bondScore calculates a composite ranking score
func bondScore(b model.Bond) float64 {
	score := 0.0
	if b.Yield != nil {
		score += *b.Yield * 3
	}
	if b.VolToday > 0 {
		score += math.Log10(float64(b.VolToday)) * 2
	}
	if b.DaysToMaturity != nil && *b.DaysToMaturity > 0 && *b.DaysToMaturity < 1100 {
		score += 5
	}
	return score
}

// --- MOEX response helpers ---
// MOEX ISS returns: {"securities": {"columns": ["SECID", ...], "data": [["RU000...", ...], ...]}}

func extractRows(data map[string]any, blockName string) []map[string]any {
	block, ok := data[blockName]
	if !ok {
		return nil
	}

	blockMap, ok := block.(map[string]any)
	if !ok {
		return nil
	}

	colsRaw, ok := blockMap["columns"].([]any)
	if !ok {
		return nil
	}
	dataRaw, ok := blockMap["data"].([]any)
	if !ok {
		return nil
	}

	columns := make([]string, len(colsRaw))
	for i, c := range colsRaw {
		columns[i], _ = c.(string)
	}

	rows := make([]map[string]any, 0, len(dataRaw))
	for _, rowRaw := range dataRaw {
		rowArr, ok := rowRaw.([]any)
		if !ok || len(rowArr) != len(columns) {
			continue
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = rowArr[i]
		}
		rows = append(rows, row)
	}

	return rows
}

func safeFirst(rows []map[string]any) map[string]any {
	if len(rows) == 0 {
		return nil
	}
	return rows[0]
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return n
		case json.Number:
			f, _ := n.Float64()
			return f
		}
	}
	return 0
}

func getFloatPtr(m map[string]any, key string) *float64 {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return &n
		case json.Number:
			f, _ := n.Float64()
			return &f
		}
	}
	return nil
}

func getInt(m map[string]any, key string) int {
	return int(getFloat(m, key))
}

func getIntPtr(m map[string]any, key string) *int {
	f := getFloatPtr(m, key)
	if f == nil {
		return nil
	}
	i := int(*f)
	return &i
}

func getInt64(m map[string]any, key string) int64 {
	return int64(getFloat(m, key))
}

func getInt64Ptr(m map[string]any, key string) *int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	i := int64(f)
	return &i
}

func safeFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func safeInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
