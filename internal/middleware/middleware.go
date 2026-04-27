package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nvaditya/forge-backend/internal/utils"
)

type contextKey string

const authClaimsContextKey contextKey = "auth_claims"

type Claims struct {
	UserID    int64
	Email     string
	Subject   string
	TokenID   string
	TokenType string
	ExpiresAt time.Time
}

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(authClaimsContextKey).(Claims)
	return claims, ok
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeJSONError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseAndValidateToken(tokenString, am.jwtSecret)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if claims.TokenType == utils.TokenTypeRefresh {
			writeJSONError(w, http.StatusUnauthorized, "refresh token cannot be used for this endpoint")
			return
		}

		if utils.IsTokenRevoked(claims.ID) {
			writeJSONError(w, http.StatusUnauthorized, "token has been revoked")
			return
		}

		requestClaims := Claims{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Subject:   claims.Subject,
			TokenID:   claims.ID,
			TokenType: claims.TokenType,
		}
		if claims.ExpiresAt != nil {
			requestClaims.ExpiresAt = claims.ExpiresAt.Time.UTC()
		}
		ctx := context.WithValue(r.Context(), authClaimsContextKey, requestClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
