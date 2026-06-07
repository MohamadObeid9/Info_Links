package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"infolinks-backend/internal/config"
	"infolinks-backend/internal/database"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/seo"
	"infolinks-backend/internal/service"
)

func testSEORouter(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{
		Port:               "8080",
		AppEnv:             "test",
		CorsAllowedOrigins: "http://localhost:8080",
		SiteBaseURL:        "http://localhost:8080",
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          "test-secret",
	}

	apiHandler := testHandler(t)

	var seoHandler *seo.Handler
	if cfg.DatabaseURL != "" {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		dbClient, err := database.New(cfg.DatabaseURL, logger)
		if err != nil {
			t.Fatalf("database.New: %v", err)
		}
		t.Cleanup(func() { _ = dbClient.Close() })
		seoService := service.NewSEOService(repository.NewPostgresSEORepository(dbClient.DB))
		seoHandler = seo.NewHandler(logger, seoService, cfg.SiteBaseURL)
	} else {
		seoHandler = seo.NewHandler(slog.Default(), nil, cfg.SiteBaseURL)
	}

	return NewRouter(cfg, apiHandler, seoHandler)
}

func TestRouterRobotsTxt(t *testing.T) {
	handler := testSEORouter(t)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("robots.txt: status %d body %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Sitemap:") {
		t.Errorf("robots.txt body: %s", rr.Body.String())
	}
}

func TestRouterSitemapXml(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	handler := testSEORouter(t)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sitemap.xml: status %d body %s", rr.Code, rr.Body.String()[:min(200, rr.Body.Len())])
	}
	if !strings.Contains(rr.Body.String(), "<urlset") {
		t.Errorf("sitemap body: %s", rr.Body.String()[:min(200, rr.Body.Len())])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
