package service

import (
	"hot-contracts-backend/model"
	"sort"
)

// SortContracts 在后端根据指定字段和顺序对合约切片进行原地排序
func SortContracts(data []model.HotContract, sortBy string, order string) {
	sort.Slice(data, func(i, j int) bool {
		var a, b float64
		var isString bool
		var sA, sB string

		switch sortBy {
		case "symbol":
			sA = data[i].Symbol
			sB = data[j].Symbol
			isString = true
		case "price":
			a = data[i].Price
			b = data[j].Price
		case "priceChangePercent":
			a = data[i].PriceChangePercent
			b = data[j].PriceChangePercent
		case "relativeStrength":
			a = data[i].RelativeStrength
			b = data[j].RelativeStrength
		case "amplitude24h":
			a = data[i].Amplitude24h
			b = data[j].Amplitude24h
		case "premiumRate":
			a = data[i].PremiumRate
			b = data[j].PremiumRate
		case "takerBuyRatio":
			a = data[i].TakerBuyRatio
			b = data[j].TakerBuyRatio
		case "netTakerVolumeUsd":
			a = data[i].NetTakerVolumeUsd
			b = data[j].NetTakerVolumeUsd
		case "volume24hUsd":
			a = data[i].Volume24hUsd
			b = data[j].Volume24hUsd
		case "fundingRate":
			// 按年化费率排序，这样4小时周期和8小时周期的合约可以公平比较
			a = data[i].AnnualizedFundingRate
			b = data[j].AnnualizedFundingRate
		case "annualizedFundingRate":
			a = data[i].AnnualizedFundingRate
			b = data[j].AnnualizedFundingRate
		case "longShortRatio":
			a = data[i].LongShortRatio
			b = data[j].LongShortRatio
		default:
			// 默认按 Volume 降序
			a = data[i].Volume24hUsd
			b = data[j].Volume24hUsd
		}

		if isString {
			if order == "ascending" {
				return sA < sB
			}
			return sA > sB
		}

		if order == "ascending" {
			return a < b
		}
		// 默认降序
		return a > b
	})
}