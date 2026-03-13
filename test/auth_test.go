package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tradeapp/server"
)

func TestAuthHandlers_DBUnavailable(t *testing.T) {
	t.Setenv("DATABASE_URL", "bad_dsn")

	srv := server.Server()
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	t.Run("login method not allowed", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/auth/login")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("login db unavailable", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/auth/login", "application/json", strings.NewReader(`{"login":"u","password":"p"}`))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
		}
	})

	t.Run("refresh method not allowed", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/auth/refresh")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("refresh db unavailable", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/auth/refresh", nil)
		if err != nil {
			t.Fatalf("build request failed: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
		}
	})

	t.Run("logout method not allowed", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/auth/logout")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("logout db unavailable", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/auth/logout", nil)
		if err != nil {
			t.Fatalf("build request failed: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
		}
	})
}
