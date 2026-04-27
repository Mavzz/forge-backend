package repository

import (
	"context"

	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
)

type StreakRepository interface {
	CreateStreak(ctx context.Context, userID int64) (sqlcdb.Streak, error)
	GetStreakByUserID(ctx context.Context, userID int64) (sqlcdb.Streak, error)
	MarkStreakCompletedToday(ctx context.Context, userID int64) (sqlcdb.Streak, error)
	ResetDailyCompletionFlags(ctx context.Context) error
}

type streakRepository struct {
	queries sqlcdb.Querier
}

func NewStreakRepository(queries sqlcdb.Querier) StreakRepository {
	return &streakRepository{queries: queries}
}

func (r *streakRepository) CreateStreak(ctx context.Context, userID int64) (sqlcdb.Streak, error) {
	return r.queries.CreateStreak(ctx, userID)
}

func (r *streakRepository) GetStreakByUserID(ctx context.Context, userID int64) (sqlcdb.Streak, error) {
	return r.queries.GetStreakByUserID(ctx, userID)
}

func (r *streakRepository) MarkStreakCompletedToday(ctx context.Context, userID int64) (sqlcdb.Streak, error) {
	return r.queries.MarkStreakCompletedToday(ctx, userID)
}

func (r *streakRepository) ResetDailyCompletionFlags(ctx context.Context) error {
	return r.queries.ResetDailyCompletionFlags(ctx)
}
