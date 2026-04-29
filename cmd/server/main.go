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

	//Setup jwt secret
	if err := api.SetJWTSecret(cfg.JWTSecret); err != nil {
		logger.Error("failed to configure JWT secret", "error", err)
		os.Exit(1)
	}

	api.SetLogger(logger.With("component", "api"))

	// Setup router
	handler := api.NewRouter(cfg)
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
