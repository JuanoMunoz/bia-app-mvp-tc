package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	Port          string
	MigrationsSQL string
	GeminiAPIKey  string
	GeminiModel   string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	config := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		MigrationsSQL: os.Getenv("MIGRATIONS_SQL"),
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
		GeminiModel:   os.Getenv("GEMINI_MODEL"),
	}
	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.MigrationsSQL == "" {
		config.MigrationsSQL = "migrations"
	}
	if config.GeminiModel == "" {
		config.GeminiModel = "gemini-2.0-flash"
	}
	return config, nil
}
