package service

import "nla/internal/model"

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
