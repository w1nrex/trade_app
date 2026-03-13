package test

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"tradeapp/steam"
)

func TestTerminal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/get/") {
			http.NotFound(w, r)
			return
		}
		encoded := strings.TrimPrefix(r.URL.Path, "/get/")
		decodedBytes, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			http.Error(w, "bad url", http.StatusBadRequest)
			return
		}
		decoded := string(decodedBytes)

		switch {
		case strings.Contains(decoded, "/inventory/"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `
			{"assets":
			[{
			"appid":730,
			"contextid":"2",
			"assetid":"1",
			"classid":"100",
			"instanceid":"0",
			"amount":"1"
			}],
			"descriptions":
			[{
			"classid":"100"
			,"instanceid":"0",
			"market_name":"AK-47 | Redline",
			"market_hash_name":"AK-47 | Redline (Field-Tested)",
			"name":"AK-47 | Redline",
			"type":"Rifle",
			"tradable":1,
			"marketable":1
			}],
			"success":1,
			"total_inventory_count":1
			}`)
		case strings.Contains(decoded, "market/priceoverview"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
			"success":true,
			"lowest_price":"$1.00",
			"volume":"100",
			"median_price":"$1.20"
			}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	fmt.Println("🔥 СЕРВЕР ЗАПУЩЕН НА:", ts.URL)
	fmt.Println("🔥 BaseURL установлен в:", ts.URL)

	steam.BaseURL = ts.URL

	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()
	r, _, err := os.Pipe()
	if err != nil {
		log.Fatalf("Не удалось создать pipe: %s", err)
	}
	os.Stdin = r
	done := make(chan struct{})
	StteamID := "76561199159548432"
	if err := steam.MainSteam(StteamID, func() {
		fmt.Println("🔄 Повторный анализ рынка...")
		close(done)
	}); err != nil {
		t.Fatalf("MainSteam error: %v", err)
	}
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("таймаут ожидания завершения")
	}
}
