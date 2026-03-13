package steam

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	requestCount = 0
	requestMutex = &sync.Mutex{}
	batchSize    = 20
	batchDelay   = 10 * time.Second
	requestDelay = 500 * time.Millisecond
)

type InventoryResponse struct {
	Assets       []Asset       `json:"assets"`
	Descriptions []Description `json:"descriptions"`
	Success      int           `json:"success"`
	TotalCount   int           `json:"total_inventory_count"`
}

type Asset struct {
	AppID      int    `json:"appid"`
	ContextID  string `json:"contextid"`
	AssetID    string `json:"assetid"`
	ClassID    string `json:"classid"`
	InstanceID string `json:"instanceid"`
	Amount     string `json:"amount"`
}

type Description struct {
	ClassID        string `json:"classid"`
	InstanceID     string `json:"instanceid"`
	MarketName     string `json:"market_name"`
	MarketHashName string `json:"market_hash_name"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Tradable       int    `json:"tradable"`
	Marketable     int    `json:"marketable"`
}

type Skins struct {
	Index          int
	Name           string
	Price          string
	MarketHashName string
	Type           string
}

func ResetBatchCounter() {
	requestMutex.Lock()
	defer requestMutex.Unlock()
	requestCount = 0
}

func Get(url string) ([]byte, error) {
	requestMutex.Lock()
	requestCount++
	currentRequest := requestCount
	requestMutex.Unlock()

	if currentRequest > 0 {
		fmt.Printf("[Запрос %d в батче]\n", currentRequest)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Bad request get: %s", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := client.Do(req)
	if err != nil && resp == nil {
		log.Printf("Bad request do: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Bad read body: %s", err)
		return nil, err
	}
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, parseErr := strconv.Atoi(retryAfter)
			if parseErr == nil {
				fmt.Printf("⚠️  Rate limit достигнут! Ждем %d секунд перед повтором...\n", seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
				ResetBatchCounter()
				return Get(url)
			}
		}
		fmt.Println("⚠️  Rate limit достигнут. Ждем 5 секунд перед повтором...")
		time.Sleep(5 * time.Second)
		ResetBatchCounter()
		return Get(url)
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("статус %d: %s", resp.StatusCode, string(body[:200]))
	}

	if currentRequest < batchSize {
		time.Sleep(requestDelay)
	} else if currentRequest == batchSize {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("⏸️  Батч из %d запросов завершен\n", batchSize)
		fmt.Printf("💤 Пауза %v перед следующим батчем...\n", batchDelay)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		time.Sleep(batchDelay)
		ResetBatchCounter()
	}

	return body, nil
}

var BaseURL = "http://127.0.0.1:8080"

func Post(url string) ([]byte, error) {
	requestMutex.Lock()
	requestCount++
	currentRequest := requestCount
	requestMutex.Unlock()

	if currentRequest > 0 {
		fmt.Printf("[Запрос %d в батче]\n", currentRequest)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	urL := BaseURL + "/get/" + url
	req, err := http.NewRequest("POST", urL, nil)
	if err != nil {
		log.Printf("Bad request post: %v", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := client.Do(req)
	if err != nil && resp == nil {
		log.Printf("Bad newrequest post : %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Bad read body: %s", err)
		return nil, err
	}
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, parseErr := strconv.Atoi(retryAfter)
			if parseErr == nil {
				fmt.Printf("⚠️  Rate limit достигнут! Ждем %d секунд перед повтором...\n", seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
				ResetBatchCounter()
				return Get(url)
			}
		}
		fmt.Println("⚠️  Rate limit достигнут. Ждем 5 секунд перед повтором...")
		time.Sleep(5 * time.Second)
		ResetBatchCounter()
		return Post(url)
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("статус %d: %s", resp.StatusCode, string(body[:200]))
	}

	if currentRequest < batchSize {
		time.Sleep(requestDelay)
	} else if currentRequest == batchSize {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("⏸️  Батч из %d запросов завершен\n", batchSize)
		fmt.Printf("💤 Пауза %v перед следующим батчем...\n", batchDelay)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		time.Sleep(batchDelay)
		ResetBatchCounter()
	}

	return body, nil

}
