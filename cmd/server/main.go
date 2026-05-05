package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/config"
	"infolinks-backend/internal/database"
)

func main() {
	//Loading Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Logger
	logger := newLogger(cfg.AppEnv)

	// Initialize database
	if err := database.InitDB(logger.With("component", "database"), cfg.DatabaseURL); err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer database.DB.Close()

	//Set Hanlder
	apiHandler := api.NewHandler(api.Dependencies{
		Logger:    logger.With("component", "api"),
		JWTSecret: []byte(cfg.JWTSecret),
	})

	// Setup router
	handler := api.NewRouter(cfg, apiHandler)
	logger.Info("backend is starting", "env", cfg.AppEnv, "port", cfg.Port)

	//Setup server
	addr := ":" + cfg.Port
	if err = http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

func newLogger(env string) *slog.Logger {
	var logHandler slog.Handler
	if env == "development" {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(logHandler).With("env", env)
	return logger
}

// func validateJWTSecret(secret string) error {
// 	secret = strings.TrimSpace(secret)
// 	if secret == "" {
// 		return errors.New("JWT_SECRET is required")
// 	}
// 	jwtSecret = []byte(secret)
// 	return nil
// }
