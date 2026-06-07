package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	AppEnv             string
	DatabaseURL        string
	JWTSecret          string
	CorsAllowedOrigins string
	SiteBaseURL        string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:               getenv("PORT", "8080"),
		AppEnv:             getenv("APP_ENV", "development"),
		CorsAllowedOrigins: getenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080,http://localhost:5173"),
		SiteBaseURL:        getenv("SITE_BASE_URL", "http://localhost:8080"),
		DatabaseURL:        getenv("DATABASE_URL"),
		JWTSecret:          getenv("JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("database_url is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt_secret is required")
	}

	return cfg, nil
}

func getenv(key string, fallback ...string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	if len(fallback) > 0 {
		return strings.TrimSpace(fallback[0])
	}
	return ""
}
