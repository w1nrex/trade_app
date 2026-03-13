package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"tradeapp/db"
	"tradeapp/steam"
)

var accessTokenSecret = []byte(getEnv("ACCESS_TOKEN_SECRET", "dev-access-secret-change-me"))

func Server() *http.Server {
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	Register(mux)
	return s
}

func Register(mux *http.ServeMux) {
	authDB, err := db.ConnectionDB()
	if err != nil {
		log.Printf("auth db disabled: %v", err)
	}

	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Foo handler")
	})
	mux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Привет %s", html.EscapeString(r.URL.Path))
	})
	mux.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
		url := strings.TrimPrefix(r.URL.Path, "/get/")
		decoded, err := base64.URLEncoding.DecodeString(url)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		url = string(decoded)
		if url == "" {
			http.Error(w, "expect /get/<url> in get handler", http.StatusBadRequest)
			return
		}
		data, err := steam.Get(url)
		if err != nil {
			log.Printf("Bad request get: %v", err)
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if authDB == nil || err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		req.Login = strings.TrimSpace(req.Login)
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password are required", http.StatusBadRequest)
			return
		}

		user, getErr := db.GetUserByLogin(authDB, req.Login)
		ip := clientIP(r)
		ua := r.UserAgent()
		if getErr != nil {
			if getErr == sql.ErrNoRows {
				_ = db.LogAuthEvent(authDB, nil, req.Login, "login_failed", ip, ua, false)
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if user.Password != req.Password {
			uid := user.ID
			_ = db.LogAuthEvent(authDB, &uid, req.Login, "login_failed", ip, ua, false)
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		refreshToken, tokenErr := generateRandomToken(32)
		if tokenErr != nil {
			http.Error(w, "token generation failed", http.StatusInternalServerError)
			return
		}
		hash := db.HashToken(refreshToken)
		refreshTTL := 7 * 24 * time.Hour
		expiresAt := time.Now().Add(refreshTTL)

		sessionID, sessErr := db.CreateAuthSession(authDB, user.ID, hash, ip, ua, expiresAt)
		if sessErr != nil {
			http.Error(w, "session creation failed", http.StatusInternalServerError)
			return
		}

		accessToken, accessErr := generateAccessToken(user.ID, sessionID, 15*time.Minute)
		if accessErr != nil {
			http.Error(w, "access token generation failed", http.StatusInternalServerError)
			return
		}

		setRefreshCookie(w, refreshToken, expiresAt)
		uid := user.ID
		_ = db.LogAuthEvent(authDB, &uid, req.Login, "login_success", ip, ua, true)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"expires_in":   900,
		})
	})

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if authDB == nil || err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		cookie, cookieErr := r.Cookie("refresh_token")
		if cookieErr != nil || cookie.Value == "" {
			http.Error(w, "missing refresh token", http.StatusUnauthorized)
			return
		}

		oldHash := db.HashToken(cookie.Value)
		ip := clientIP(r)
		ua := r.UserAgent()
		session, sessErr := db.GetActiveAuthSessionByHash(authDB, oldHash, time.Now())
		if sessErr != nil {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}

		newRefreshToken, tokenErr := generateRandomToken(32)
		if tokenErr != nil {
			http.Error(w, "token generation failed", http.StatusInternalServerError)
			return
		}
		newHash := db.HashToken(newRefreshToken)
		newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

		userID, newSessionID, rotateErr := db.RotateAuthSession(authDB, oldHash, newHash, ip, ua, newExpiresAt)
		if rotateErr != nil {
			http.Error(w, "refresh failed", http.StatusUnauthorized)
			return
		}

		accessToken, accessErr := generateAccessToken(userID, newSessionID, 15*time.Minute)
		if accessErr != nil {
			http.Error(w, "access token generation failed", http.StatusInternalServerError)
			return
		}

		setRefreshCookie(w, newRefreshToken, newExpiresAt)
		_ = db.LogAuthEvent(authDB, &session.UserID, "", "token_refresh", ip, ua, true)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"expires_in":   900,
		})
	})

	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if authDB == nil || err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		if cookie, cookieErr := r.Cookie("refresh_token"); cookieErr == nil && cookie.Value != "" {
			hash := db.HashToken(cookie.Value)
			_ = db.RevokeAuthSessionByHash(authDB, hash)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   isCookieSecure(),
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
		w.WriteHeader(http.StatusNoContent)
	})
}

// Генерация токена
func generateRandomToken(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Генерация access токена с пользовательскими данными и временем жизни
func generateAccessToken(userID, sessionID int64, ttl time.Duration) (string, error) {
	type claims struct {
		Sub int64 `json:"sub"`
		Sid int64 `json:"sid"`
		Iat int64 `json:"iat"`
		Exp int64 `json:"exp"`
	}

	now := time.Now().Unix()
	payload, err := json.Marshal(claims{
		Sub: userID,
		Sid: sessionID,
		Iat: now,
		Exp: now + int64(ttl.Seconds()),
	})
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(payloadB64)
	return payloadB64 + "." + sig, nil
}

// Подпись токена для обеспечения его целостности и защиты от подделки
func sign(payload string) string {
	h := hmac.New(sha256.New, accessTokenSecret)
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// Настройка cookie для хранения refresh токена с безопасными атрибутами и временем жизни
func setRefreshCookie(w http.ResponseWriter, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func isCookieSecure() bool {
	return strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true")
}

// Получение IP-адреса клиента из заголовков или удаленного адреса для логирования и безопасности
func clientIP(r *http.Request) string {
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
