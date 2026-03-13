package steam

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

func Url(steamID string, markethashname string) string {
	var urlStr string
	if steamID == "" && markethashname != "" {
		encodedName := url.QueryEscape(markethashname)
		urlStr = fmt.Sprintf(
			"https://steamcommunity.com/market/priceoverview/?appid=730&currency=1&market_hash_name=%s",
			encodedName)
	} else if steamID != "" && markethashname == "" {
		urlStr = fmt.Sprintf("https://steamcommunity.com/inventory/%s/730/2", steamID)
	} else if steamID != "" && markethashname != "" {
		log.Println(" ⚠️  Введите либо SteamID, либо MarketHashName, но не оба одновременно.")
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(urlStr))
	return encoded
}

func GetInventory(steamID string) (*InventoryResponse, error) {
	url := Url(steamID, "")
	body, err := Post(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения инвентаря: %v", err)
	}

	var inventory InventoryResponse
	if err := json.Unmarshal(body, &inventory); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	return &inventory, nil
}
