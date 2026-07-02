package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"hot-contracts-backend/model"

	"github.com/go-resty/resty/v2"
)

const (
	fapiBaseUrl    = "https://fapi.binance.com"
	apiBaseUrl     = "https://api.binance.com"
	dataBaseUrl    = "https://www.binance.com" // 数据接口使用 www 域名
)

type BinanceService struct {
	client *resty.Client
}

func NewBinanceService() *BinanceService {
	client := resty.New()
	// Binance API 有时响应较慢，特别是在国内网络环境下，增加超时时间
	client.SetTimeout(15 * time.Second)
	// 增加重试次数和重试间隔
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	// 设置 User-Agent，否则 /futures/data/* 接口会返回 403
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	return &BinanceService{
		client: client,
	}
}

// FetchAllData 并发获取所有需要的原始数据
func (s *BinanceService) FetchAllData() ([]model.RawTicker, []model.RawSpotTicker, []model.RawPremiumIndex, error) {
	var (
		wg             sync.WaitGroup
		tickers        []model.RawTicker
		spotTickers    []model.RawSpotTicker
		premiumIndices []model.RawPremiumIndex
		errTickers     error
		errSpot        error
		errPremium     error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		resp, err := s.client.R().Get(fapiBaseUrl + "/fapi/v1/ticker/24hr")
		if err != nil {
			errTickers = err
			return
		}
		errTickers = json.Unmarshal(resp.Body(), &tickers)
	}()

	go func() {
		defer wg.Done()
		resp, err := s.client.R().Get(apiBaseUrl + "/api/v3/ticker/price")
		if err != nil {
			errSpot = err
			return
		}
		errSpot = json.Unmarshal(resp.Body(), &spotTickers)
	}()

	go func() {
		defer wg.Done()
		resp, err := s.client.R().Get(fapiBaseUrl + "/fapi/v1/premiumIndex")
		if err != nil {
			errPremium = err
			return
		}
		errPremium = json.Unmarshal(resp.Body(), &premiumIndices)
	}()

	wg.Wait()

	if errTickers != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch tickers: %w", errTickers)
	}
	if errSpot != nil {
		log.Printf("Warning: failed to fetch spot tickers: %v", errSpot)
	}
	if errPremium != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch premium indices: %w", errPremium)
	}

	return tickers, spotTickers, premiumIndices, nil
}

// fetchLongShortRatiosForSymbols 并发获取指定合约的多空账户比
func (s *BinanceService) fetchLongShortRatiosForSymbols(symbols []string) map[string]float64 {
	result := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 限制并发数，避免触发API限流
	sem := make(chan struct{}, 10)

	for _, symbol := range symbols {
		wg.Add(1)
		sem <- struct{}{}

		go func(sym string) {
			defer wg.Done()
			defer func() { <-sem }()

			url := fmt.Sprintf("%s/futures/data/globalLongShortAccountRatio?symbol=%s&period=1h&limit=1", dataBaseUrl, sym)
			resp, err := s.client.R().Get(url)
			if err != nil {
				return
			}

			var ratios []model.RawLSRatio
			if err := json.Unmarshal(resp.Body(), &ratios); err != nil {
				return
			}

			if len(ratios) > 0 {
				if ratio, err := strconv.ParseFloat(ratios[0].LongShortRatio, 64); err == nil {
					mu.Lock()
					result[sym] = ratio
					mu.Unlock()
				}
			}
		}(symbol)
	}

	wg.Wait()
	return result
}

// fetchTakerRatiosForSymbols 并发获取指定合约的主动买卖比
func (s *BinanceService) fetchTakerRatiosForSymbols(symbols []string) map[string]float64 {
	result := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 10)

	for _, symbol := range symbols {
		wg.Add(1)
		sem <- struct{}{}

		go func(sym string) {
			defer wg.Done()
			defer func() { <-sem }()

			url := fmt.Sprintf("%s/futures/data/takerlongshortRatio?symbol=%s&period=1h&limit=1", dataBaseUrl, sym)
			resp, err := s.client.R().Get(url)
			if err != nil {
				log.Printf("Warning: failed to fetch taker ratio for %s: %v", sym, err)
				return
			}

			if resp.StatusCode() != 200 {
				log.Printf("Warning: taker ratio API returned %d for %s", resp.StatusCode(), sym)
				return
			}

			var ratios []model.RawTakerRatio
			if err := json.Unmarshal(resp.Body(), &ratios); err != nil {
				return
			}

			if len(ratios) > 0 {
				if ratio, err := strconv.ParseFloat(ratios[0].BuySellRatio, 64); err == nil {
					mu.Lock()
					result[sym] = ratio
					mu.Unlock()
				}
			}
		}(symbol)
	}

	wg.Wait()
	return result
}

// fetchFundingInfo 获取所有合约的资金费率结算周期
func (s *BinanceService) fetchFundingInfo() map[string]int {
	result := make(map[string]int)

	resp, err := s.client.R().Get(fapiBaseUrl + "/fapi/v1/fundingInfo")
	if err != nil {
		log.Printf("Warning: failed to fetch fundingInfo: %v", err)
		return result
	}

	var infos []model.RawFundingInfo
	if err := json.Unmarshal(resp.Body(), &infos); err != nil {
		log.Printf("Warning: failed to parse fundingInfo: %v", err)
		return result
	}

	for _, info := range infos {
		result[info.Symbol] = info.FundingIntervalHours
	}

	return result
}

// parseFloat is a helper to parse float strings safely
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// BuildHotContracts 聚合数据并推导高阶指标
func (s *BinanceService) BuildHotContracts() ([]model.HotContract, error) {
	tickers, spotTickers, premiumIndices, err := s.FetchAllData()
	if err != nil {
		return nil, err
	}

	// 获取资金费率结算周期
	fundingInfoMap := s.fetchFundingInfo()

	// 1. 建立 Lookup Maps
	spotMap := make(map[string]float64)
	for _, st := range spotTickers {
		spotMap[st.Symbol] = parseFloat(st.Price)
	}

	premiumMap := make(map[string]model.RawPremiumIndex)
	for _, pi := range premiumIndices {
		premiumMap[pi.Symbol] = pi
	}

	// 2. 提取 BTC 基准涨跌幅 (用于相对强弱)
	var btcChangePercent float64
	for _, t := range tickers {
		if t.Symbol == "BTCUSDT" {
			btcChangePercent = parseFloat(t.PriceChangePercent)
			break
		}
	}

	// 3. 先过滤出活跃合约，再获取多空比（减少API调用）
	var activeSymbols []string
	for _, t := range tickers {
		if len(t.Symbol) < 4 || t.Symbol[len(t.Symbol)-4:] != "USDT" {
			continue
		}
		quoteVol := parseFloat(t.QuoteVolume)
		if quoteVol < 1000000 {
			continue
		}
		activeSymbols = append(activeSymbols, t.Symbol)
	}

	// 并发获取活跃合约的多空比和主动买卖比
	longShortMap := s.fetchLongShortRatiosForSymbols(activeSymbols)
	takerRatioMap := s.fetchTakerRatiosForSymbols(activeSymbols)

	// 4. Hash Join 与推导
	var contracts []model.HotContract

	for _, t := range tickers {
		// 只看 USDT 本位合约
		if len(t.Symbol) < 4 || t.Symbol[len(t.Symbol)-4:] != "USDT" {
			continue
		}

		price := parseFloat(t.LastPrice)
		quoteVol := parseFloat(t.QuoteVolume)

		// 过滤掉交易额过小的不活跃合约 (例如 < 100万 U)
		if quoteVol < 1000000 {
			continue
		}

		highPrice := parseFloat(t.HighPrice)
		lowPrice := parseFloat(t.LowPrice)
		priceChangePct := parseFloat(t.PriceChangePercent)

		// 基础衍生计算
		amplitude := 0.0
		if lowPrice > 0 {
			amplitude = (highPrice - lowPrice) / lowPrice
		}

		// 主动买卖比 (来自 takerlongshortRatio API)
		// buySellRatio 是 buy/sell，需要转换为 buy/(buy+sell)
		buySellRatio := takerRatioMap[t.Symbol]
		takerBuyRatio := 0.0
		if buySellRatio > 0 {
			takerBuyRatio = buySellRatio / (1 + buySellRatio)
		}
		// 计算净流入额: (buyRatio - sellRatio) * quoteVol = (buyRatio - (1-buyRatio)) * quoteVol
		netTakerVolume := (takerBuyRatio - 0.5) * 2 * quoteVol

		// Spot 数据
		spotPrice := spotMap[t.Symbol]
		premiumRate := 0.0
		if spotPrice > 0 {
			premiumRate = (price - spotPrice) / spotPrice
		}

		// Premium/Funding 数据
		var markPriceDeviation, fundingRate, annualizedFundingRate float64
		var nextFundingTime int64
		fundingIntervalHours := 8 // 默认 8 小时

		// 从 fundingInfo 获取准确的结算周期
		if hours, ok := fundingInfoMap[t.Symbol]; ok && hours > 0 {
			fundingIntervalHours = hours
		}

		if pi, ok := premiumMap[t.Symbol]; ok {
			markPrice := parseFloat(pi.MarkPrice)
			if markPrice > 0 {
				markPriceDeviation = (price - markPrice) / markPrice
			}
			fundingRate = parseFloat(pi.LastFundingRate)
			annualizedFundingRate = fundingRate * (24.0 / float64(fundingIntervalHours)) * 365
			nextFundingTime = pi.NextFundingTime
		}

		// 多空比
		longShortRatio := longShortMap[t.Symbol]

		contract := model.HotContract{
			Symbol:                t.Symbol,
			Price:                 price,
			SpotPrice:             spotPrice,
			MarkPriceDeviation:    markPriceDeviation,
			PremiumRate:           premiumRate,
			PriceChangePercent:    priceChangePct,
			RelativeStrength:      priceChangePct - btcChangePercent,
			Amplitude24h:          amplitude,
			Volume24hUsd:          quoteVol,
			TakerBuyRatio:         takerBuyRatio,
			NetTakerVolumeUsd:     netTakerVolume,
			FundingRate:           fundingRate,
			FundingIntervalHours:  fundingIntervalHours,
			AnnualizedFundingRate: annualizedFundingRate,
			NextFundingTime:       nextFundingTime,
			LongShortRatio:        longShortRatio,
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}
