package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"infolinks-backend/internal/config"
)

func TestRouterMetrics_openInDevelopment(t *testing.T) {
	cfg := config.Config{
		AppEnv:    "development",
		JWTSecret: "test-secret",
	}
	handler := NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), testHandler(t), nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "go_goroutines") {
		t.Fatalf("body missing prometheus metrics: %q", rr.Body.String())
	}
}

func TestRouterMetrics_requiresAuthWhenConfigured(t *testing.T) {
	cfg := config.Config{
		AppEnv:               "development",
		JWTSecret:            "test-secret",
		MetricsBasicAuthUser: "scraper",
		MetricsBasicAuthPass: "secret",
	}
	handler := NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), testHandler(t), nil)

	t.Run("missing credentials", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("valid credentials", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("scraper", "secret")
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "go_goroutines") {
			t.Fatalf("body missing prometheus metrics")
		}
	})
}

func TestRouterMetrics_deniedInProductionWithoutAuthConfig(t *testing.T) {
	cfg := config.Config{
		AppEnv:    "production",
		JWTSecret: "test-secret",
	}
	handler := NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), testHandler(t), nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}
