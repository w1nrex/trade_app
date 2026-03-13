package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func ConnectionDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=server dbname=pqgo sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

type MarketAnalysis struct {
	ItemName         string  `json:"item_name"`
	LowestPrice      float64 `json:"lowest_price"`
	MedianPrice      float64 `json:"median_price"`
	Volume           int     `json:"volume"`
	RecommendedPrice float64 `json:"recommended_price"`
	Liquidity        string  `json:"liquidity"`
	PriceSpread      float64 `json:"price_spread"`
}

func InsertDB(analysis MarketAnalysis) func(db *sql.DB) {
	return func(db *sql.DB) {
		_, _ = db.Exec(
			"INSERT INTO skins (item_name, lowest_price, median_price, volume, recommended_price, liquidity, price_spread) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			analysis.ItemName,
			analysis.LowestPrice,
			analysis.MedianPrice,
			analysis.Volume,
			analysis.RecommendedPrice,
			analysis.Liquidity,
			analysis.PriceSpread,
		)
	}
}

type User struct {
	ID       int64
	Login    string
	Nickname string
	Password string
	SteamID  int64
}

type AuthSession struct {
	ID               int64
	UserID           int64
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        sql.NullTime
}

func UserExistsByLogin(db *sql.DB, login string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM "user" WHERE login = $1)`, login).Scan(&exists)
	return exists, err
}

func GetUserByLogin(db *sql.DB, login string) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, login, nickname, password, steam_id FROM "user" WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.Nickname, &user.Password, &user.SteamID)
	return user, err
}

func CheckUser(db *sql.DB, login, password string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM "user" WHERE login = $1 AND password = $2)`,
		login,
		password,
	).Scan(&exists)
	return exists, err
}

func CreateUser(db *sql.DB, login, nickname, password string, steamID int64) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO "user" (login, nickname, password, steam_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		login,
		nickname,
		password,
		steamID,
	).Scan(&id)
	return id, err
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func CreateAuthSession(db *sql.DB, userID int64, refreshTokenHash, ip, userAgent string, expiresAt time.Time) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO auth_session (user_id, refresh_token_hash, ip, user_agent, expires_at, last_used_at)
		 VALUES ($1, $2, NULLIF($3, '')::inet, $4, $5, now())
		 RETURNING id`,
		userID,
		refreshTokenHash,
		ip,
		userAgent,
		expiresAt,
	).Scan(&id)
	return id, err
}

func GetActiveAuthSessionByHash(db *sql.DB, refreshTokenHash string, now time.Time) (AuthSession, error) {
	var s AuthSession
	err := db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, expires_at, revoked_at
		 FROM auth_session
		 WHERE refresh_token_hash = $1
		   AND revoked_at IS NULL
		   AND expires_at > $2`,
		refreshTokenHash,
		now,
	).Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.ExpiresAt, &s.RevokedAt)
	return s, err
}

func RotateAuthSession(db *sql.DB, oldHash, newHash, ip, userAgent string, newExpiresAt time.Time) (int64, int64, error) {
	var userID int64
	err := db.QueryRow(
		`UPDATE auth_session
		 SET revoked_at = now(), last_used_at = now()
		 WHERE refresh_token_hash = $1
		   AND revoked_at IS NULL
		 RETURNING user_id`,
		oldHash,
	).Scan(&userID)
	if err != nil {
		return 0, 0, err
	}

	newSessionID, err := CreateAuthSession(db, userID, newHash, ip, userAgent, newExpiresAt)
	if err != nil {
		return 0, 0, err
	}
	return userID, newSessionID, nil
}

func RevokeAuthSessionByHash(db *sql.DB, refreshTokenHash string) error {
	_, err := db.Exec(
		`UPDATE auth_session
		 SET revoked_at = now(), last_used_at = now()
		 WHERE refresh_token_hash = $1
		   AND revoked_at IS NULL`,
		refreshTokenHash,
	)
	return err
}

func LogAuthEvent(db *sql.DB, userID *int64, login, eventType, ip, userAgent string, success bool) error {
	if len(eventType) > 32 {
		return fmt.Errorf("event type is too long")
	}
	_, err := db.Exec(
		`INSERT INTO auth_event (user_id, login, event_type, ip, user_agent, success)
		 VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, '')::inet, $5, $6)`,
		userID,
		login,
		eventType,
		ip,
		userAgent,
		success,
	)
	return err
}
