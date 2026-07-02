package model

// HotContract 是发送给 Vue 的最终归一化数据结构
type HotContract struct {
	Symbol string `json:"symbol"` // e.g., "BTCUSDT"

	// --- 价格、基差与波动指标 ---
	Price              float64 `json:"price"`
	SpotPrice          float64 `json:"spotPrice"`
	MarkPriceDeviation float64 `json:"markPriceDeviation"` // (LastPrice - MarkPrice) / MarkPrice
	PremiumRate        float64 `json:"premiumRate"`        // (Futures - Spot) / Spot
	PriceChangePercent float64 `json:"priceChangePercent"`
	RelativeStrength   float64 `json:"relativeStrength"`   // CoinChange% - BTCChange%
	Amplitude24h       float64 `json:"amplitude24h"`

	// --- 资金流向与动能指标 ---
	Volume24hUsd      float64 `json:"volume24hUsd"`
	TakerBuyRatio     float64 `json:"takerBuyRatio"`       // TakerBuy / TotalVolume (e.g., 55.2%)
	NetTakerVolumeUsd float64 `json:"netTakerVolumeUsd"`
	OpenInterestUsd   float64 `json:"openInterestUsd"`
	OiChangePercent24 float64 `json:"oiChangePercent24h"`
	VolToOiRatio      float64 `json:"volToOiRatio"`

	// --- 情绪与套利指标 ---
	FundingRate           float64 `json:"fundingRate"`
	FundingIntervalHours  int     `json:"fundingIntervalHours"`  // 资金费率结算周期（小时）
	AnnualizedFundingRate float64 `json:"annualizedFundingRate"`
	NextFundingTime       int64   `json:"nextFundingTime"`   // 毫秒时间戳，前端计算倒计时
	LongShortRatio        float64 `json:"longShortRatio"`
}