package repository

import (
	"context"

	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg sqlcdb.CreateUserParams) (sqlcdb.User, error)
	GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlcdb.User, error)
	UpdateUserStreaks(ctx context.Context, arg sqlcdb.UpdateUserStreaksParams) (sqlcdb.User, error)
}

type userRepository struct {
	queries sqlcdb.Querier
}

func NewUserRepository(queries sqlcdb.Querier) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) CreateUser(ctx context.Context, arg sqlcdb.CreateUserParams) (sqlcdb.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (sqlcdb.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *userRepository) UpdateUserStreaks(ctx context.Context, arg sqlcdb.UpdateUserStreaksParams) (sqlcdb.User, error) {
	return r.queries.UpdateUserStreaks(ctx, arg)
}
