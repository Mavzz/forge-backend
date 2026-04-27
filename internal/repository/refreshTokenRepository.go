package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
)

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (sqlcdb.RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenHash string) (sqlcdb.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokensByUser(ctx context.Context, userID int64) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
}

type refreshTokenRepository struct {
	queries sqlcdb.Querier
}

func NewRefreshTokenRepository(queries sqlcdb.Querier) RefreshTokenRepository {
	return &refreshTokenRepository{queries: queries}
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (sqlcdb.RefreshToken, error) {
	return r.queries.CreateRefreshToken(ctx, sqlcdb.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt.UTC(), Valid: true},
	})
}

func (r *refreshTokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (sqlcdb.RefreshToken, error) {
	return r.queries.GetRefreshToken(ctx, tokenHash)
}

func (r *refreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

func (r *refreshTokenRepository) RevokeAllRefreshTokensByUser(ctx context.Context, userID int64) error {
	return r.queries.RevokeAllRefreshTokensByUser(ctx, userID)
}

func (r *refreshTokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return r.queries.DeleteExpiredRefreshTokens(ctx)
}
