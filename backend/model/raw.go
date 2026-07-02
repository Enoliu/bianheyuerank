package model

// RawTicker 对应 /fapi/v1/ticker/24hr
type RawTicker struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	QuoteVolume        string `json:"quoteVolume"`
	TakerBuyQuoteVol   string `json:"takerBuyQuoteVolume"`
}

// RawSpotTicker 对应现货 /api/v3/ticker/price
type RawSpotTicker struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// RawPremiumIndex 对应 /fapi/v1/premiumIndex
type RawPremiumIndex struct {
	Symbol          string `json:"symbol"`
	MarkPrice       string `json:"markPrice"`
	LastFundingRate string `json:"lastFundingRate"`
	NextFundingTime int64  `json:"nextFundingTime"`
	Time            int64  `json:"time"`
}

// RawOpenInterest 对应 /fapi/v1/openInterestHist
type RawOpenInterest struct {
	Symbol               string `json:"symbol"`
	SumOpenInterestValue string `json:"sumOpenInterestValue"`
}

// RawLSRatio 对应 /futures/data/globalLongShortAccountRatio
type RawLSRatio struct {
	Symbol         string `json:"symbol"`
	LongShortRatio string `json:"longShortRatio"`
}

// RawTakerRatio 对应 /futures/data/takerlongshortRatio
type RawTakerRatio struct {
	Symbol         string `json:"symbol"`
	BuySellRatio   string `json:"buySellRatio"`
	BuyVol         string `json:"buyVol"`
	SellVol        string `json:"sellVol"`
	Timestamp      int64  `json:"timestamp"`
}

// RawFundingInfo 对应 /fapi/v1/fundingInfo
type RawFundingInfo struct {
	Symbol                string `json:"symbol"`
	AdjustedFundingRateCap   string `json:"adjustedFundingRateCap"`
	AdjustedFundingRateFloor string `json:"adjustedFundingRateFloor"`
	FundingIntervalHours  int    `json:"fundingIntervalHours"`
}