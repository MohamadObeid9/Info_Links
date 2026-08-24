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
			if cc := rr.Header().Get("Cache-Control"); cc != cacheControlHTML {
				t.Fatalf("Cache-Control: got %q want %q", cc, cacheControlHTML)
			}
		})
	}
}

func TestStaticCacheControl(t *testing.T) {
	tests := []struct {
		path string
		html bool
		want string
	}{
		{path: "/", html: true, want: cacheControlHTML},
		{path: "/about", html: true, want: cacheControlHTML},
		{path: "/assets/index-CQUV9458.js", want: cacheControlHashed},
		{path: "/assets/index-DpG80_j0.css", want: cacheControlHashed},
		{path: "/assets/inter-latin-800-normal-BYj_oED-.woff2", want: cacheControlHashed},
		{path: "/assets/favicon-32x32.png", want: cacheControlStatic},
		{path: "/registerSW.js", want: cacheControlStatic},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := staticCacheControl(tt.path, tt.html)
			if got != tt.want {
				t.Fatalf("staticCacheControl(%q, %v) = %q, want %q", tt.path, tt.html, got, tt.want)
			}
		})
	}
}

func TestStaticHandler_unhashedAssetCacheControl(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/assets/favicon-32x32.png", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != cacheControlStatic {
		t.Fatalf("Cache-Control: got %q want %q", cc, cacheControlStatic)
	}
}
