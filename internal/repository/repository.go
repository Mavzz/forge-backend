package repository

import sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"

type Repositories struct {
	Users         UserRepository
	Tasks         TaskRepository
	Categories    CategoryRepository
	Streaks       StreakRepository
	RefreshTokens RefreshTokenRepository
}

func NewRepositories(queries sqlcdb.Querier) *Repositories {
	return &Repositories{
		Users:         NewUserRepository(queries),
		Tasks:         NewTaskRepository(queries),
		Categories:    NewCategoryRepository(queries),
		Streaks:       NewStreakRepository(queries),
		RefreshTokens: NewRefreshTokenRepository(queries),
	}
}
