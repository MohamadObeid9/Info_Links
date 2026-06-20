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

func testStaticRouter(t *testing.T) http.Handler {
	t.Helper()
	t.Chdir("../..")
	cfg := config.Config{
		AppEnv:    "test",
		JWTSecret: "test-secret",
	}
	return NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), testHandler(t), nil)
}

func TestStaticHandler_probePathsReturn404(t *testing.T) {
	handler := testStaticRouter(t)

	for _, path := range []string{
		"/inputs.php",
		"/wp-admin/includes/index.php",
		"/wp-theme.php",
		"/random-garbage",
		"/assets/missing-file.js",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusNotFound {
				t.Fatalf("status: got %d want %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}

func TestStaticHandler_spaPathsReturn200(t *testing.T) {
	handler := testStaticRouter(t)

	for _, path := range []string{
		"/",
		"/feedback",
		"/about",
		"/report-submit",
		"/admin-gate",
		"/admin",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
			}
			if !strings.Contains(rr.Body.String(), "<html") {
				t.Fatalf("expected index.html body for SPA path %q", path)
			}
		})
	}
}
