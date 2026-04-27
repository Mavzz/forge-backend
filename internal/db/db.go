package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvaditya/forge-backend/internal/config"
)

func ConnectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database config: %w", err)
	}

	port, err := strconv.ParseUint(cfg.DB.Port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid database port %q: %w", cfg.DB.Port, err)
	}

	poolConfig.ConnConfig.Host = cfg.DB.Host
	poolConfig.ConnConfig.Port = uint16(port)
	poolConfig.ConnConfig.Database = cfg.DB.Name
	poolConfig.ConnConfig.User = cfg.DB.User
	poolConfig.ConnConfig.Password = cfg.DB.Password
	poolConfig.ConnConfig.TLSConfig = nil

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
