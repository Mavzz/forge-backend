package repository

import (
	"context"

	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, arg sqlcdb.CreateCategoryParams) (sqlcdb.Category, error)
	ListCategories(ctx context.Context) ([]sqlcdb.Category, error)
	GetCategoryByID(ctx context.Context, id int64) (sqlcdb.Category, error)
}

type categoryRepository struct {
	queries sqlcdb.Querier
}

func NewCategoryRepository(queries sqlcdb.Querier) CategoryRepository {
	return &categoryRepository{queries: queries}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, arg sqlcdb.CreateCategoryParams) (sqlcdb.Category, error) {
	return r.queries.CreateCategory(ctx, arg)
}

func (r *categoryRepository) ListCategories(ctx context.Context) ([]sqlcdb.Category, error) {
	return r.queries.ListCategories(ctx)
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id int64) (sqlcdb.Category, error) {
	return r.queries.GetCategoryByID(ctx, id)
}
