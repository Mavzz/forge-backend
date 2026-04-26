package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
	"github.com/nvaditya/forge-backend/internal/models"
	"github.com/nvaditya/forge-backend/internal/repository"
	"github.com/nvaditya/forge-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

var repos *repository.Repositories
var jwtSigningKey []byte

func SetRepositories(r *repository.Repositories) {
	repos = r
}

func SetJWTSecret(secret string) {
	jwtSigningKey = []byte(secret)
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

	token, err := utils.CreateAccessToken(user.ID, userUUID, user.Email, jwtSigningKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(models.TokenResponse{
		AccessToken: token,
		UUID:        userUUID,
	})

}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "logout is not implemented yet")
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
