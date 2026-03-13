package steam

import (
	"encoding/json"
	"fmt"
)

type PriceOverview struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}

func FetchSkinPrice(marketHashName string) (*PriceOverview, error) {
	url := Url("", marketHashName)
	body, err := Post(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}

	var priceData PriceOverview
	if err := json.Unmarshal(body, &priceData); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	if !priceData.Success {
		return nil, fmt.Errorf("цена не найдена (success=false)")
	}

	if priceData.LowestPrice == "" {
		return nil, fmt.Errorf("пустая цена")
	}

	return &priceData, nil
}
