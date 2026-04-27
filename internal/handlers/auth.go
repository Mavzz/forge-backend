package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
	"github.com/nvaditya/forge-backend/internal/middleware"
	"github.com/nvaditya/forge-backend/internal/models"
	"github.com/nvaditya/forge-backend/internal/repository"
	"github.com/nvaditya/forge-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

var repos *repository.Repositories
var jwtSigningKey []byte
var dbPool *pgxpool.Pool

func SetRepositories(r *repository.Repositories) {
	repos = r
}

func SetJWTSecret(secret string) {
	jwtSigningKey = []byte(secret)
}

func SetDB(pool *pgxpool.Pool) {
	dbPool = pool
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if repos == nil || len(jwtSigningKey) == 0 {
		writeJSONError(w, http.StatusInternalServerError, "server dependencies are not initialized")
		return
	}

	var req models.SignupRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)

	if username == "" || email == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}
	if len(password) < 8 {
		writeJSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid email address")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	createdUser, err := repos.Users.CreateUser(r.Context(), sqlcdb.CreateUserParams{
		Name:         username,
		Email:        email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSONError(w, http.StatusConflict, "user with this email already exists")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if !createdUser.Uuid.Valid {
		writeJSONError(w, http.StatusInternalServerError, "failed to load user uuid")
		return
	}
	createdUserUUID := uuid.UUID(createdUser.Uuid.Bytes).String()

	token, err := utils.CreateAccessToken(createdUser.ID, createdUserUUID, createdUser.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(models.SignupResponse{
		ID:       createdUser.ID,
		Username: createdUser.Name,
		Email:    createdUser.Email,
		Token:    token,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if repos == nil || len(jwtSigningKey) == 0 {
		writeJSONError(w, http.StatusInternalServerError, "server dependencies are not initialized")
		return
	}

	var req models.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)

	if email == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if _, err := mail.ParseAddress(email); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid email address")
		return
	}

	user, err := repos.Users.GetUserByEmail(r.Context(), email)

	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !user.Uuid.Valid {
		writeJSONError(w, http.StatusInternalServerError, "failed to load user uuid")
		return
	}
	userUUID := uuid.UUID(user.Uuid.Bytes).String()

	accessToken, err := utils.CreateAccessToken(user.ID, userUUID, user.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := utils.CreateRefreshToken(user.ID, userUUID, user.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshClaims, err := utils.ParseAndValidateToken(refreshToken, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshExpiry := time.Now().UTC().Add(utils.RefreshTokenTTL)
	if refreshClaims.ExpiresAt != nil {
		refreshExpiry = refreshClaims.ExpiresAt.Time.UTC()
	}

	if _, err := repos.RefreshTokens.CreateRefreshToken(r.Context(), user.ID, hashToken(refreshToken), refreshExpiry); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UUID:         userUUID,
	})

}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	if repos == nil || len(jwtSigningKey) == 0 || dbPool == nil {
		writeJSONError(w, http.StatusInternalServerError, "server dependencies are not initialized")
		return
	}

	var req models.RefreshTokenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	incomingToken := strings.TrimSpace(req.RefreshToken)
	if incomingToken == "" {
		writeJSONError(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	claims, err := utils.ParseAndValidateToken(incomingToken, jwtSigningKey)
	if err != nil || claims.TokenType != utils.TokenTypeRefresh {
		writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	tokenHash := hashToken(incomingToken)
	stored, err := repos.RefreshTokens.GetRefreshToken(r.Context(), tokenHash)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if stored.Revoked {
		writeJSONError(w, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	if stored.ExpiresAt.Valid && stored.ExpiresAt.Time.UTC().Before(time.Now().UTC()) {
		writeJSONError(w, http.StatusUnauthorized, "refresh token has expired")
		return
	}

	user, err := repos.Users.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if !user.Uuid.Valid {
		writeJSONError(w, http.StatusInternalServerError, "failed to load user uuid")
		return
	}
	userUUID := uuid.UUID(user.Uuid.Bytes).String()

	if claims.Subject != "" && claims.Subject != userUUID {
		writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	newAccessToken, err := utils.CreateAccessToken(user.ID, userUUID, user.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	newRefreshToken, err := utils.CreateRefreshToken(user.ID, userUUID, user.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	newRefreshClaims, err := utils.ParseAndValidateToken(newRefreshToken, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	newExpiry := time.Now().UTC().Add(utils.RefreshTokenTTL)
	if newRefreshClaims.ExpiresAt != nil {
		newExpiry = newRefreshClaims.ExpiresAt.Time.UTC()
	}

	// Perform insert + revoke atomically in a single transaction so that token
	// rotation cannot leave the database in a partially-updated state.
	tx, err := dbPool.Begin(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to start session")
		return
	}
	defer func() {
		if rbErr := tx.Rollback(r.Context()); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			log.Printf("token rotation: failed to rollback transaction: %v", rbErr)
		}
	}()

	txRepo := repository.NewRefreshTokenRepository(sqlcdb.New(tx))

	if _, err := txRepo.CreateRefreshToken(r.Context(), user.ID, hashToken(newRefreshToken), newExpiry); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save session")
		return
	}

	if err := txRepo.RevokeRefreshToken(r.Context(), tokenHash); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to rotate session")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to rotate session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(models.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		UUID:         userUUID,
	})
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	if repos == nil || len(jwtSigningKey) == 0 {
		writeJSONError(w, http.StatusInternalServerError, "server dependencies are not initialized")
		return
	}

	authClaims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || authClaims.UserID <= 0 {
		writeJSONError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// Revoke the access token in-memory so it is rejected immediately for its remaining TTL.
	utils.RevokeToken(authClaims.TokenID, authClaims.ExpiresAt)

	var req models.RefreshTokenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil && !errors.Is(err, io.EOF) {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	incomingRefresh := strings.TrimSpace(req.RefreshToken)
	if incomingRefresh != "" {
		refreshClaims, parseErr := utils.ParseAndValidateToken(incomingRefresh, jwtSigningKey)
		if parseErr != nil || refreshClaims.TokenType != utils.TokenTypeRefresh {
			writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}

		if refreshClaims.UserID != authClaims.UserID {
			writeJSONError(w, http.StatusForbidden, "refresh token does not belong to authenticated user")
			return
		}

		if err := repos.RefreshTokens.RevokeRefreshToken(r.Context(), hashToken(incomingRefresh)); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to revoke session")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
