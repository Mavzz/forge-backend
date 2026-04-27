package utils

import (
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

	revokedTokensMu.Lock()
	defer revokedTokensMu.Unlock()

	expiry, exists := revokedTokenIDs[tokenID]
	if !exists {
		return false
	}

	if !expiry.After(now) {
		delete(revokedTokenIDs, tokenID)
		return false
	}

	return true
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
