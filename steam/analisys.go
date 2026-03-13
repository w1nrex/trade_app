package steam

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"tradeapp/db"
)

type MarketAnalysis struct {
	ItemName         string  `json:"item_name"`
	LowestPrice      float64 `json:"lowest_price"`
	MedianPrice      float64 `json:"median_price"`
	Volume           int     `json:"volume"`
	RecommendedPrice float64 `json:"recommended_price"`
	Liquidity        string  `json:"liquidity"`
	PriceSpread      float64 `json:"price_spread"`
}

func parsePrice(priceStr string) float64 {
	var price float64
	priceStr = strings.TrimPrefix(priceStr, "$")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	price, _ = strconv.ParseFloat(priceStr, 64)
	return price
}

func parseVolume(volumeStr string) int {
	volumeStr = strings.TrimPrefix(volumeStr, "$")
	volumeStr = strings.ReplaceAll(volumeStr, ",", "")
	volume, _ := strconv.ParseFloat(volumeStr, 64)
	return int(volume)
}

func Analysis(skins []Skins) (db.MarketAnalysis, error) {
	var result db.MarketAnalysis
	totalBatches := (len(skins) + batchSize - 1) / batchSize

	fmt.Println("\n=== АНАЛИЗ РЫНКА ===")
	fmt.Printf("Всего скинов: %d\n", len(skins))
	fmt.Printf("Батчей: %d (по %d скинов)\n\n", totalBatches, batchSize)

	for _, skin := range skins {
		price, err := FetchSkinPrice(skin.MarketHashName)
		if err != nil {
			log.Printf("Ошибка в получении данных %s: %v", skin.Name, err)
		}
		lowest := parsePrice(price.LowestPrice)
		median := parsePrice(price.MedianPrice)
		volume := parseVolume(price.Volume)

		if lowest == 0 {
			return db.MarketAnalysis{}, fmt.Errorf("Недостаточно данных для анализа.")
		}
		if median == 0 {
			return db.MarketAnalysis{}, fmt.Errorf("Недостаточно данных для анализа.")
		}

		var volumeFactor float64
		switch {
		case volume >= 1000:
			volumeFactor = 0.7
		case volume >= 500:
			volumeFactor = 0.5
		case volume >= 100:
			volumeFactor = 0.3
		default:
			volumeFactor = 1.0
		}

		marketPrice := (lowest * (1 - volumeFactor)) + (median * volumeFactor)
		priceSpread := math.Abs(median-lowest) / lowest
		if priceSpread > 0.2 {
			marketPrice *= 0.95
		}
		spread := 0.0
		if lowest > 0 {
			spread = math.Abs(median-lowest) / lowest * 100
		}

		analys := db.MarketAnalysis{
			ItemName:         skin.MarketHashName,
			LowestPrice:      lowest,
			MedianPrice:      median,
			Volume:           volume,
			RecommendedPrice: marketPrice,
			Liquidity:        getLiquidity(volume),
			PriceSpread:      math.Round(spread*100) / 100,
		}
		result = analys
	}
	return result, nil
}

func getLiquidity(volume int) string {
	switch {
	case volume >= 1000:
		return "Высокая 🟢"
	case volume >= 500:
		return "Средняя 🟡"
	case volume >= 100:
		return "Низкая 🟠"
	default:
		return "Очень низкая 🔴"
	}
}
