package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/config"
	"infolinks-backend/internal/database"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/seo"
	"infolinks-backend/internal/service"
	"infolinks-backend/internal/webbotauth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.AppEnv, cfg.LogLevel)

	dbClient, err := database.New(cfg.DatabaseURL, logger.With("component", "database"))
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = dbClient.Close() }()

	services, userService := handleServices(dbClient.DB)

	webBotDir, err := webbotauth.NewDirectory(cfg.JWTSecret, cfg.SiteBaseURL)
	if err != nil {
		logger.Error("web bot auth directory failed", "error", err)
		os.Exit(1)
	}

	apiHandler, err := api.NewHandler(api.Dependencies{
		DB:                  dbClient,
		JWTSecret:           []byte(cfg.JWTSecret),
		SiteBaseURL:         cfg.SiteBaseURL,
		SupabaseURL:         cfg.SupabaseURL,
		SupabaseAnonKey:     cfg.SupabaseAnonKey,
		WebBotAuth:          webBotDir,
		UserService:         services.UserService,
		AnalyticsService:    services.AnalyticsService,
		LinkService:         services.LinkService,
		CourseService:       services.CourseService,
		ReportService:       services.ReportService,
		ContentService:      services.ContentService,
		FeedbackService:     services.FeedbackService,
		PageViewService:     services.PageViewService,
		LinkClickService:    services.LinkClickService,
		ExtraLinkService:    services.ExtraLinkService,
		ContributionService: services.ContributionService,
		ExtraSectionService: services.ExtraSectionService,
		ServiceService:      services.ServiceService,
		Logger:              logger.With("component", "api"),
	})
	if err != nil {
		logger.Error("api handler initialization failed", "error", err)
		os.Exit(1)
	}

	seoHandler := seo.NewHandler(
		logger.With("component", "seo"),
		service.NewSEOService(repository.NewPostgresSEORepository(dbClient.DB)),
		cfg.SiteBaseURL,
	)

	handler := api.NewRouter(
		cfg,
		logger.With("component", "http"),
		apiHandler,
		seoHandler,
	)

	startStaleGuestCleanup(userService, logger.With("component", "guest-cleanup"))

	logger.Info("backend is starting", "env", cfg.AppEnv, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err = http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// startStaleGuestCleanup deletes unclaimed guests idle for StaleGuestTTL, once
// at boot and then hourly. Cascaded analytics for those guests go with them.
func startStaleGuestCleanup(svc *service.UserService, logger *slog.Logger) {
	const every = time.Hour

	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		n, err := svc.DeleteStaleGuests(ctx, service.StaleGuestTTL)
		if err != nil {
			logger.Error("stale guest cleanup failed", "error", err)
			return
		}
		if n > 0 {
			logger.Info("deleted stale unclaimed guests", "count", n, "ttl", service.StaleGuestTTL.String())
		}
	}

	run()
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}

func newLogger(appEnv, logLevel string) *slog.Logger {
	var logHandler slog.Handler
	if appEnv == "development" {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else if logLevel == "info" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler).With("AppEnv", appEnv)
	return logger
}

func handleServices(db *sql.DB) (*api.Dependencies, *service.UserService) {

	userRepo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo)

	analyticsRepo := repository.NewPostgresAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo)

	linkRepo := repository.NewPostgresLinkRepository(db)
	linkService := service.NewLinkService(linkRepo)

	courseRepo := repository.NewPostgresCourseRepository(db)
	courseService := service.NewCourseService(courseRepo)

	reportRepo := repository.NewPostgresReportRepository(db)
	reportService := service.NewReportService(reportRepo)

	feedbackRepo := repository.NewPostgresFeedbackRepository(db)
	feedbackService := service.NewFeedbackService(feedbackRepo)

	contentRepo := repository.NewPostgresContentRepository(db)
	contentService := service.NewContentService(contentRepo)

	pageViewRepo := repository.NewPostgresPageViewRepository(db)
	pageViewService := service.NewPageViewService(pageViewRepo)

	linkClickRepo := repository.NewPostgresLinkClickRepository(db)
	linkClickService := service.NewLinkClickService(linkClickRepo)

	contributionsRepo := repository.NewPostgresContributionRepository(db)
	contributionsService := service.NewContributionService(contributionsRepo)

	extraSectionRepo := repository.NewPostgresExtraSectionRepository(db)
	extraSectionService := service.NewExtraSectionService(extraSectionRepo)

	extraLinkRepo := repository.NewPostgresExtraLinkRepository(db)
	extraLinkService := service.NewExtraLinkService(extraLinkRepo)

	serviceRepo := repository.NewPostgresServiceRepository(db)
	serviceService := service.NewServiceService(serviceRepo)

	return &api.Dependencies{
		UserService:         userService,
		AnalyticsService:    analyticsService,
		LinkService:         linkService,
		CourseService:       courseService,
		ReportService:       reportService,
		ContentService:      contentService,
		FeedbackService:     feedbackService,
		PageViewService:     pageViewService,
		LinkClickService:    linkClickService,
		ContributionService: contributionsService,
		ExtraSectionService: extraSectionService,
		ExtraLinkService:    extraLinkService,
		ServiceService:      serviceService,
	}, userService
}
