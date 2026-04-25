package repository

import (
	"context"

	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, arg sqlcdb.CreateTaskParams) (sqlcdb.Task, error)
	ListTasksByUser(ctx context.Context, userID int64) ([]sqlcdb.Task, error)
	CompleteTask(ctx context.Context, id int64) (sqlcdb.Task, error)
	DeleteTask(ctx context.Context, id int64) error
}

type taskRepository struct {
	queries sqlcdb.Querier
}

func NewTaskRepository(queries sqlcdb.Querier) TaskRepository {
	return &taskRepository{queries: queries}
}

func (r *taskRepository) CreateTask(ctx context.Context, arg sqlcdb.CreateTaskParams) (sqlcdb.Task, error) {
	return r.queries.CreateTask(ctx, arg)
}

func (r *taskRepository) ListTasksByUser(ctx context.Context, userID int64) ([]sqlcdb.Task, error) {
	return r.queries.ListTasksByUser(ctx, userID)
}

func (r *taskRepository) CompleteTask(ctx context.Context, id int64) (sqlcdb.Task, error) {
	return r.queries.CompleteTask(ctx, id)
}

func (r *taskRepository) DeleteTask(ctx context.Context, id int64) error {
	return r.queries.DeleteTask(ctx, id)
}