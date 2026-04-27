package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

const (
	accessTokenTTL  = 24 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email,omitempty"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

var (
	revokedTokensMu sync.RWMutex
	revokedTokenIDs = make(map[string]time.Time)
)

func CreateAccessToken(userID int64, userUUID, email string, jwtSigningKey []byte) (string, error) {
	return createToken(userID, userUUID, email, TokenTypeAccess, accessTokenTTL, jwtSigningKey)
}

func CreateRefreshToken(userID int64, userUUID, email string, jwtSigningKey []byte) (string, error) {
	return createToken(userID, userUUID, email, TokenTypeRefresh, RefreshTokenTTL, jwtSigningKey)
}

func ParseAndValidateToken(tokenString string, jwtSigningKey []byte) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return jwtSigningKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.UserID <= 0 {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func RevokeToken(tokenID string, expiry time.Time) {
	if tokenID == "" {
		return
	}

	if expiry.IsZero() {
		expiry = time.Now().UTC().Add(RefreshTokenTTL)
	} else {
		expiry = expiry.UTC()
	}

	revokedTokensMu.Lock()
	revokedTokenIDs[tokenID] = expiry
	revokedTokensMu.Unlock()
}

func IsTokenRevoked(tokenID string) bool {
	if tokenID == "" {
		return false
	}

	now := time.Now().UTC()

	// Use a read lock for the common (non-expired) path to reduce contention.
	revokedTokensMu.RLock()
	expiry, exists := revokedTokenIDs[tokenID]
	revokedTokensMu.RUnlock()

	if !exists {
		return false
	}

	if expiry.After(now) {
		return true
	}

	// Entry has expired – promote to write lock to delete it (re-check to avoid race).
	revokedTokensMu.Lock()
	expiry, exists = revokedTokenIDs[tokenID]
	if exists && !expiry.After(now) {
		delete(revokedTokenIDs, tokenID)
	}
	revokedTokensMu.Unlock()

	return false
}

// StartTokenCleanup starts a background goroutine that periodically removes
// expired entries from the in-memory revoked-token map to prevent unbounded growth.
func StartTokenCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				purgeExpiredTokens()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func purgeExpiredTokens() {
	now := time.Now().UTC()
	revokedTokensMu.Lock()
	defer revokedTokensMu.Unlock()
	for id, expiry := range revokedTokenIDs {
		if !expiry.After(now) {
			delete(revokedTokenIDs, id)
		}
	}
}

func createToken(userID int64, userUUID, email, tokenType string, ttl time.Duration, jwtSigningKey []byte) (string, error) {
	now := time.Now().UTC()
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSigningKey)
}
