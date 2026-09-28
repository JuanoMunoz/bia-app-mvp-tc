package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	Port           string
	MigrationsSQL  string
	GeminiAPIKey   string
	GeminiModel    string
	AllowedOrigins []string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	config := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		Port:           os.Getenv("PORT"),
		MigrationsSQL:  os.Getenv("MIGRATIONS_SQL"),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		GeminiModel:    os.Getenv("GEMINI_MODEL"),
		AllowedOrigins: parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
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

// parseAllowedOrigins interpreta ALLOWED_ORIGINS como lista separada por comas.
// Acepta "*" como comodín (ej. "https://*.vercel.app" cubre los deployments de preview).
// Sin variable, conserva los orígenes de desarrollo local.
func parseAllowedOrigins(raw string) []string {
	defaults := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	if strings.TrimSpace(raw) == "" {
		return defaults
	}
	origins := make([]string, 0)
	seen := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		origin := strings.TrimSuffix(strings.TrimSpace(part), "/")
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; !ok {
			seen[origin] = struct{}{}
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return defaults
	}
	return origins
}
