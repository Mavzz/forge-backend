package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	APIVersion string
	DB         DBConfig
}

type DBConfig struct {
	User     string
	Host     string
	Name     string
	Password string
	Port     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	config := &Config{
		Port:       os.Getenv("PORT"),
		APIVersion: os.Getenv("API_VERSION"),
		DB: DBConfig{
			User:     os.Getenv("PG_USER"),
			Host:     os.Getenv("PG_HOST"),
			Name:     os.Getenv("PG_DB"),
			Password: os.Getenv("PG_PASSWORD"),
			Port:     os.Getenv("PG_PORT"),
		},
	}

	return config, nil
}
