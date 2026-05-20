package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/config"
	"infolinks-backend/internal/database"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.AppEnv)

	dbClient, err := database.New(cfg.DatabaseURL, logger.With("component", "database"))
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer dbClient.Close()

	services := handleServices(dbClient.DB)
	apiHandler, err := api.NewHandler(api.Dependencies{
		Logger:              logger.With("component", "api"),
		JWTSecret:           []byte(cfg.JWTSecret),
		ReportService:       services.ReportService,
		ContributionService: services.ContributionService,
	})
	if err != nil {
		logger.Error("api handler initialization failed", "error", err)
		os.Exit(1)
	}

	handler := api.NewRouter(cfg, apiHandler)
	logger.Info("backend is starting", "env", cfg.AppEnv, "port", cfg.Port)

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

func handleServices(db *sql.DB) *api.Dependencies {
	reportRepo := repository.NewPostgresReportRepository(db)
	reportService := service.NewReportService(reportRepo)

	contributionsRepo := repository.NewPostgresContributionRepository(db)
	contributionsService := service.NewContributionService(contributionsRepo)

	dependencies := &api.Dependencies{
		ReportService:       reportService,
		ContributionService: contributionsService,
	}

	return dependencies
}
