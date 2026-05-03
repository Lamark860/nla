package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"nla/internal/model"
)

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
