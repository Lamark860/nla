package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

	secids := make([]string, len(bonds))
	for i, b := range bonds {
		secids[i] = b.SECID
	}

	issuerMap, err := s.issuerRepo.GetBySecids(ctx, secids)
	if err != nil {
		return nil, fmt.Errorf("fetch issuer mappings: %w", err)
	}

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
			if g.emitterName == "" && issuer.EmitterName != "" {
				g.emitterName = issuer.EmitterName
			}
			g.bonds = append(g.bonds, b)
		} else {
			noIssuerBonds = append(noIssuerBonds, b)
		}
	}

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

	if len(noIssuerBonds) > 0 {
		result = append(result, model.IssuerGroup{
			EmitterID:   0,
			EmitterName: "Без эмитента",
			BondCount:   len(noIssuerBonds),
			Bonds:       noIssuerBonds,
		})
	}

	// Sort by credit rating (highest first), then bond count.
	// Uses ScoreOrd (1-22 normalized scale) so cross-agency comparison is correct —
	// the legacy Score (1-10) collapses BBB- and BB+ into the same bucket.
	ratingScoreMap := make(map[int64]int)
	if s.ratingRepo != nil {
		allRatings, _ := s.ratingRepo.GetAll(ctx)
		for _, r := range allRatings {
			if r.ScoreOrd > ratingScoreMap[r.EmitterID] {
				ratingScoreMap[r.EmitterID] = r.ScoreOrd
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		si := ratingScoreMap[result[i].EmitterID]
		sj := ratingScoreMap[result[j].EmitterID]
		if si != sj {
			return si > sj
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

	if n, _ := s.redis.Del(ctx, bondsCacheKey).Result(); n > 0 {
		total += n
	}

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
