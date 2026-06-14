package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	AppEnv               string
	JWTSecret            string
	DatabaseURL          string
	SiteBaseURL          string
	SupabaseURL          string
	SupabaseAnonKey      string
	CorsAllowedOrigins   string
	MetricsBasicAuthUser string
	MetricsBasicAuthPass string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:                 getenv("PORT", "8080"),
		AppEnv:               getenv("APP_ENV", "development"),
		CorsAllowedOrigins:   getenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080,http://localhost:5173"),
		SiteBaseURL:          getenv("SITE_BASE_URL", "http://localhost:8080"),
		DatabaseURL:          getenv("DATABASE_URL"),
		JWTSecret:            getenv("JWT_SECRET"),
		SupabaseURL:          getenv("SUPABASE_URL"),
		SupabaseAnonKey:      getenv("SUPABASE_ANON_KEY"),
		MetricsBasicAuthUser: getenv("METRICS_BASIC_AUTH_USER"),
		MetricsBasicAuthPass: getenv("METRICS_BASIC_AUTH_PASSWORD"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("database url is required")
	}

	if cfg.SupabaseURL == "" {
		return Config{}, fmt.Errorf("supabase url is required")
	}

	if cfg.SupabaseAnonKey == "" {
		return Config{}, fmt.Errorf("supabase anon key is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt secret is required")
	}

	if err := cfg.validateMetricsAuth(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) MetricsAuthEnabled() bool {
	return c.MetricsBasicAuthUser != "" && c.MetricsBasicAuthPass != ""
}

func (c Config) validateMetricsAuth() error {
	hasUser := c.MetricsBasicAuthUser != ""
	hasPass := c.MetricsBasicAuthPass != ""

	if hasUser != hasPass {
		return fmt.Errorf("metrics basic auth user and password must both be set")
	}

	if c.AppEnv == "production" && !c.MetricsAuthEnabled() {
		return fmt.Errorf("metrics basic auth is required in production")
	}

	return nil
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
