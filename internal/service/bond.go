package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
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
	ratingRepo *mongorepo.RatingRepo
}

func NewBondService(moexClient *moex.Client, rdb *redis.Client, issuerRepo *mongorepo.IssuerRepo, ratingRepo *mongorepo.RatingRepo) *BondService {
	return &BondService{moex: moexClient, redis: rdb, issuerRepo: issuerRepo, ratingRepo: ratingRepo}
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
	marketdataYields := extractRows(raw, "marketdata_yields")

	if len(securities) == 0 {
		return nil, fmt.Errorf("bond %s not found", secid)
	}

	bond := parseBond(securities[0], safeFirst(marketdata))
	mergeYieldData(&bond, safeFirst(marketdataYields))
	calculateFields(&bond)

	// Resolve emitter_id from bond_issuers collection
	if s.issuerRepo != nil {
		if issuer, err := s.issuerRepo.GetBySecid(ctx, secid); err == nil && issuer != nil && issuer.EmitterID > 0 {
			bond.EmitterID = &issuer.EmitterID
			bond.EmitterName = issuer.EmitterName
		}
	}

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

	// Sort by credit rating (highest first), then bond count
	// Build rating score map by emitter_id (direct lookup)
	ratingScoreMap := make(map[int64]int)
	if s.ratingRepo != nil {
		allRatings, _ := s.ratingRepo.GetAll(ctx)
		for _, r := range allRatings {
			if r.Score > ratingScoreMap[r.EmitterID] {
				ratingScoreMap[r.EmitterID] = r.Score
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		si := ratingScoreMap[result[i].EmitterID]
		sj := ratingScoreMap[result[j].EmitterID]
		if si != sj {
			return si > sj // Higher rating first
		}
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

// ClearCache removes all bond-related keys from Redis
func (s *BondService) ClearCache(ctx context.Context) (int64, error) {
	var total int64

	// Delete list cache
	if n, _ := s.redis.Del(ctx, bondsCacheKey).Result(); n > 0 {
		total += n
	}

	// Delete individual bond caches (bonds:SECID)
	var cursor uint64
	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, bondCachePrefix+"*", 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			if n, _ := s.redis.Del(ctx, keys...).Result(); n > 0 {
				total += n
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return total, nil
}

// ToggleIssuer hides/shows all bonds of an emitter
func (s *BondService) ToggleIssuer(ctx context.Context, emitterID int64, hidden bool) (int64, error) {
	return s.issuerRepo.ToggleHidden(ctx, emitterID, hidden)
}

// SyncMissingIssuers finds bonds present in MOEX but absent from bond_issuers,
// fetches their EMITTER_ID from MOEX description API, and creates bond_issuer records.
func (s *BondService) SyncMissingIssuers(ctx context.Context) (int, error) {
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return 0, fmt.Errorf("get bonds: %w", err)
	}

	existing, err := s.issuerRepo.GetAllSecids(ctx)
	if err != nil {
		return 0, fmt.Errorf("get existing secids: %w", err)
	}

	var missing []string
	for _, b := range bonds {
		if !existing[b.SECID] {
			missing = append(missing, b.SECID)
		}
	}

	if len(missing) == 0 {
		return 0, nil
	}

	log.Printf("[issuer-sync] Found %d bonds missing from bond_issuers", len(missing))

	synced := 0
	for _, secid := range missing {
		select {
		case <-ctx.Done():
			return synced, ctx.Err()
		default:
		}

		raw, err := s.moex.GetDisclosure(ctx, secid)
		if err != nil {
			log.Printf("[issuer-sync] WARN: fetch disclosure for %s: %v", secid, err)
			continue
		}

		descRows := extractRows(raw, "description")
		var emitterID int64
		for _, row := range descRows {
			if getString(row, "name") == "EMITTER_ID" {
				if v, err := strconv.ParseInt(getString(row, "value"), 10, 64); err == nil {
					emitterID = v
				}
				break
			}
		}

		issuer := &model.BondIssuer{
			SECID:     secid,
			EmitterID: emitterID,
		}
		if err := s.issuerRepo.Upsert(ctx, issuer); err != nil {
			log.Printf("[issuer-sync] WARN: upsert issuer %s: %v", secid, err)
			continue
		}
		synced++
		log.Printf("[issuer-sync] Added %s (emitter_id=%d)", secid, emitterID)

		// Small delay to avoid hammering MOEX
		time.Sleep(200 * time.Millisecond)
	}

	return synced, nil
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
	now := time.Now()

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

	// Days to maturity + years
	if b.MatDate != "" {
		if matDate, err := time.Parse("2006-01-02", b.MatDate); err == nil {
			days := int(time.Until(matDate).Hours() / 24)
			b.DaysToMaturity = &days
			years := float64(days) / 365.25
			b.YearsToMaturity = &years
		}
	}

	// Days to put option
	if b.PutOptionDate != "" {
		if d, err := time.Parse("2006-01-02", b.PutOptionDate); err == nil {
			days := int(d.Sub(now).Hours() / 24)
			b.DaysToPut = &days
			b.IsNearOffer = days >= 0 && days <= 90
		}
	}

	// Days to call option
	if b.CallOptionDate != "" {
		if d, err := time.Parse("2006-01-02", b.CallOptionDate); err == nil {
			days := int(d.Sub(now).Hours() / 24)
			b.DaysToCall = &days
		}
	}

	// Days to next coupon
	if b.NextCoupon != "" {
		if d, err := time.Parse("2006-01-02", b.NextCoupon); err == nil {
			days := int(d.Sub(now).Hours() / 24)
			b.DaysToNextCoupon = &days
		}
	}

	// Bond type flags
	bt := strings.ToLower(b.BondType)
	btnLower := strings.ToLower(b.BondTypeName)
	b.IsFloat = strings.Contains(btnLower, "флоатер") || strings.Contains(bt, "float") || strings.Contains(btnLower, "float")
	b.IsIndexed = strings.Contains(btnLower, "индексируемые") || b.BondType == "6"
	b.HasNoFixedCoupon = b.IsFloat || b.IsIndexed

	// Category — by BOARDID (like ASH), fallback to bond type string
	switch b.BoardID {
	case "TQOB": // ОФЗ
		b.BondCategory = "ОФЗ"
	case "TQCB": // Корпоративные
		b.BondCategory = "Корпоративная"
	case "TQIR": // Ипотечные
		b.BondCategory = "Ипотечная"
	default:
		switch {
		case b.CurrencyID == "USD" || b.FaceUnit == "USD":
			b.BondCategory = "Еврооблигация"
		case strings.Contains(btnLower, "субфед"):
			b.BondCategory = "Субфедеральная"
		case strings.Contains(btnLower, "муницип"):
			b.BondCategory = "Муниципальная"
		case strings.Contains(btnLower, "офз") || strings.Contains(bt, "офз"):
			b.BondCategory = "ОФЗ"
		default:
			b.BondCategory = "Корпоративная"
		}
	}

	// Coupon display
	if b.CouponPercent > 0 {
		cp := b.CouponPercent
		b.CouponDisplay = &cp
	} else if b.CouponValue > 0 && b.FaceValue > 0 {
		cd := (b.CouponValue / b.FaceValue) * 100
		b.CouponDisplay = &cd
	}

	// Current yield
	if b.CouponPercent > 0 && b.Last != nil && *b.Last > 0 {
		cy := (b.CouponPercent / *b.Last) * 100
		b.CurrentYield = &cy
	}

	// Modified duration
	if b.Duration != nil && b.Yield != nil && *b.Yield > 0 {
		md := float64(*b.Duration) / (1 + *b.Yield/100)
		b.ModifiedDuration = &md
	}

	// Bid/Offer spread
	if b.Bid != nil && b.Offer != nil {
		abs := *b.Offer - *b.Bid
		b.SpreadAbsolute = &abs
		if *b.Bid > 0 {
			pct := abs / *b.Bid * 100
			b.SpreadPercent = &pct
		}
	}

	// Mid price
	if b.Bid != nil && b.Offer != nil {
		mid := (*b.Bid + *b.Offer) / 2
		b.MidPricePct = &mid
		if b.FaceValue > 0 {
			midRub := (mid / 100) * b.FaceValue
			b.MidPriceRUB = &midRub
		}
	}

	// Bid/Offer in RUB
	if b.Bid != nil && b.FaceValue > 0 {
		bidRub := (*b.Bid / 100) * b.FaceValue
		b.BidRUB = &bidRub
	}
	if b.Offer != nil && b.FaceValue > 0 {
		offerRub := (*b.Offer / 100) * b.FaceValue
		b.OfferRUB = &offerRub
	}

	// Average trade size
	if b.NumTrades != nil && *b.NumTrades > 0 && b.ValueTodayRUB != nil {
		avg := *b.ValueTodayRUB / float64(*b.NumTrades)
		b.AvgTradeSize = &avg
	}

	// Total depth
	if b.BidDepth != nil && b.OfferDepth != nil {
		td := *b.BidDepth + *b.OfferDepth
		b.TotalDepth = &td
	}

	// Bid/offer ratio
	if b.BidDepth != nil && b.OfferDepth != nil && *b.BidDepth > 0 {
		ratio := float64(*b.OfferDepth) / float64(*b.BidDepth)
		b.BidOfferRatio = &ratio
	}

	// Accrued interest as %
	if b.AccruedInt > 0 && b.FaceValue > 0 {
		pct := b.AccruedInt / b.FaceValue * 100
		b.AccruedIntPct = &pct
	}

	// Life progress (0-100%)
	if b.SettleDate != "" && b.MatDate != "" {
		if settle, err := time.Parse("2006-01-02", b.SettleDate); err == nil {
			if mat, err := time.Parse("2006-01-02", b.MatDate); err == nil {
				total := mat.Sub(settle).Hours()
				elapsed := now.Sub(settle).Hours()
				if total > 0 {
					prog := elapsed / total * 100
					if prog < 0 {
						prog = 0
					}
					if prog > 100 {
						prog = 100
					}
					b.LifeProgress = &prog
				}
			}
		}
	}

	// Trading status text
	switch b.TradingStatus {
	case "T":
		b.TradingStatusTxt = "Торги идут"
	case "N":
		b.TradingStatusTxt = "Не ведутся"
	case "S":
		b.TradingStatusTxt = "Приостановлены"
	default:
		b.TradingStatusTxt = "Неизвестно"
	}

	// Risk category
	b.RiskCategory = calcRiskCategory(b)
}

// calcRiskCategory determines risk level based on available data
func calcRiskCategory(b *model.Bond) string {
	y := safeFloat(b.Yield)
	dtm := safeInt(b.DaysToMaturity)

	switch {
	case y > 25 || dtm <= 0:
		return "toxic"
	case y > 18:
		return "high"
	case y > 14:
		return "medium-high"
	case y > 10:
		return "medium"
	default:
		return "low"
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
			return safeFloat(bonds[i].ValueTodayRUB) > safeFloat(bonds[j].ValueTodayRUB)
		})
	case "coupon_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].CouponDisplay) > safeFloat(bonds[j].CouponDisplay)
		})
	case "coupon_asc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeFloat(bonds[i].CouponDisplay) < safeFloat(bonds[j].CouponDisplay)
		})
	case "duration_asc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeInt(bonds[i].Duration) < safeInt(bonds[j].Duration)
		})
	case "duration_desc":
		sort.Slice(bonds, func(i, j int) bool {
			return safeInt(bonds[i].Duration) > safeInt(bonds[j].Duration)
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

// mergeYieldData supplements bond with data from marketdata_yields block
// (only fills fields that are still nil — marketdata takes priority)
func mergeYieldData(b *model.Bond, yld map[string]any) {
	if yld == nil {
		return
	}
	if b.EffectiveYield == nil {
		b.EffectiveYield = getFloatPtr(yld, "EFFECTIVEYIELD")
	}
	if b.YieldAtWAPrice == nil {
		b.YieldAtWAPrice = getFloatPtr(yld, "YIELDATWAPRICE")
	}
	if b.YieldToPrevYield == nil {
		b.YieldToPrevYield = getFloatPtr(yld, "YIELDTOPREVYIELD")
	}
	if b.CloseYield == nil {
		b.CloseYield = getFloatPtr(yld, "CLOSEYIELD")
	}
	if b.ZSpread == nil {
		b.ZSpread = getFloatPtr(yld, "ZSPREAD")
	}
	if b.ZSpreadAtWAPrice == nil {
		b.ZSpreadAtWAPrice = getFloatPtr(yld, "ZSPREADATWAPRICE")
	}
	if b.IRICPIClose == nil {
		b.IRICPIClose = getFloatPtr(yld, "IRICPICLOSE")
	}
	if b.BEIClose == nil {
		b.BEIClose = getFloatPtr(yld, "BEICLOSE")
	}
	if b.CBRClose == nil {
		b.CBRClose = getFloatPtr(yld, "CBRCLOSE")
	}
	if b.Duration == nil {
		b.Duration = getIntPtr(yld, "DURATION")
	}
	if b.YieldToOffer == nil {
		b.YieldToOffer = getFloatPtr(yld, "YIELDTOOFFER")
	}
	if b.YieldLastCoupon == nil {
		b.YieldLastCoupon = getFloatPtr(yld, "YIELDLASTCOUPON")
	}
	if b.CallOptionYield == nil {
		b.CallOptionYield = getFloatPtr(yld, "CALLOPTIONYIELD")
	}
	if b.CallOptionDuration == nil {
		b.CallOptionDuration = getIntPtr(yld, "CALLOPTIONDURATION")
	}
	if b.DurationWAPrice == nil {
		b.DurationWAPrice = getIntPtr(yld, "DURATIONWAPRICE")
	}
}

// SyncMissingRatingsFromMoex fetches credit ratings from MOEX CCI API
// for emitters that exist in bond_issuers but have no records in issuer_ratings.
func (s *BondService) SyncMissingRatingsFromMoex(ctx context.Context) (int, error) {
	// Get all emitter IDs from bond_issuers
	allIssuers, err := s.issuerRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("get issuers: %w", err)
	}

	// Get emitter IDs that already have ratings
	rated, err := s.ratingRepo.GetDistinctEmitterIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("get rated emitters: %w", err)
	}

	// Find emitters without ratings
	type missingInfo struct {
		emitterID   int64
		emitterName string
	}
	seen := make(map[int64]bool)
	var missing []missingInfo
	for _, iss := range allIssuers {
		if iss.EmitterID == 0 || rated[iss.EmitterID] || seen[iss.EmitterID] {
			continue
		}
		seen[iss.EmitterID] = true
		missing = append(missing, missingInfo{emitterID: iss.EmitterID, emitterName: iss.EmitterName})
	}

	if len(missing) == 0 {
		return 0, nil
	}

	log.Printf("[rating-sync] Found %d emitters without ratings, trying MOEX CCI", len(missing))

	synced := 0
	for _, m := range missing {
		select {
		case <-ctx.Done():
			return synced, ctx.Err()
		default:
		}

		cciRatings, err := s.moex.GetCCIRatings(ctx, m.emitterID)
		if err != nil {
			log.Printf("[rating-sync] WARN: CCI fetch for emitter %d: %v", m.emitterID, err)
			continue
		}
		if len(cciRatings) == 0 {
			continue
		}

		for _, cr := range cciRatings {
			rating := &model.IssuerRating{
				EmitterID: m.emitterID,
				Issuer:    m.emitterName,
				Agency:    cr.AgencyName,
				Rating:    cr.RatingValue,
			}
			if err := s.ratingRepo.Upsert(ctx, rating); err != nil {
				log.Printf("[rating-sync] WARN: upsert rating emitter %d %s: %v", m.emitterID, cr.AgencyName, err)
			}
		}

		synced++
		log.Printf("[rating-sync] MOEX CCI: emitter %d %q → %d ratings", m.emitterID, m.emitterName, len(cciRatings))

		time.Sleep(200 * time.Millisecond)
	}

	return synced, nil
}
