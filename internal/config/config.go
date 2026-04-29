package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	AppEnv             string
	DatabaseURL        string
	JWTSecret          string
	CorsAllowedOrigins string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:               getenv("PORT", "8080"),
		AppEnv:             getenv("APP_ENV", "development"),
		CorsAllowedOrigins: getenv("CORS_ALLOWED_ORIGINS", ""),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("database_url is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt_secret is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}
