package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"tradeapp/server"
	"tradeapp/steam"
)

const SteamID = "76561199159548432"

func TestPath(t *testing.T) {
	server := server.Server()
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()

	steamUrl := steam.Url(SteamID, "")
	urL := "/get/" + steamUrl

	t.Run("url", func(t *testing.T) {
		resp, err := http.Get(ts.URL + urL)
		if err != nil {
			t.Fatalf("Bad request get: %s", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Bad read: %s", err)
		}
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		fmt.Println(result)
	})
}
