package test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tradeapp/server"
)

func TestServerStart(t *testing.T) {
	server := server.Server()
	ts := httptest.NewServer(server.Handler)
	defer ts.Close()
	t.Run("foo handler", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/foo")
		if err != nil {
			t.Fatalf("Request failded: %s", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Foo handler" {
			t.Errorf("Expected 'Foo handler, got %s", string(body))
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

	})

	t.Run("bar handler", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/bar")
		if err != nil {
			log.Fatalf("Request bar failed: %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		got := string(body)

		if !strings.HasPrefix(got, "Привет") {
			t.Errorf("Expected to start with 'Привет', got %q", got)
		}
		if !strings.Contains(got, "/bar") {
			t.Errorf("Expected to contain '/bar', got %q", got)
		}
	})
}
