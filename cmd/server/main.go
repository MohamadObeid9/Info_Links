package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/app"
	"infolinks-backend/internal/config"
	"infolinks-backend/internal/database"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/seo"
	"infolinks-backend/internal/service"
	"infolinks-backend/internal/webbotauth"
)

const shutdownTimeout = 10 * time.Second

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

	services, userService := app.Wire(dbClient.DB)

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startStaleGuestCleanup(ctx, userService, logger.With("component", "guest-cleanup"))

	server := newHTTPServer(":"+cfg.Port, handler)
	logger.Info("backend is starting", "env", cfg.AppEnv, "port", cfg.Port)
	if err := serveHTTP(ctx, server, nil); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// serveHTTP serves until ctx is cancelled, then shuts down gracefully.
// If ln is non-nil, Serve is used (tests); otherwise ListenAndServe.
func serveHTTP(ctx context.Context, server *http.Server, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		if ln != nil {
			err = server.Serve(ln)
		} else {
			err = server.ListenAndServe()
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	}
}

// startStaleGuestCleanup deletes unclaimed guests idle for StaleGuestTTL, once
// at boot and then hourly until ctx is cancelled. Cascaded analytics for those
// guests go with them.
func startStaleGuestCleanup(ctx context.Context, svc *service.UserService, logger *slog.Logger) {
	const every = time.Hour

	run := func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		n, err := svc.DeleteStaleGuests(runCtx, service.StaleGuestTTL)
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
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
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
