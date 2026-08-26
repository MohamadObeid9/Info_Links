package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestStaticHandler_markdownNegotiation(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/markdown")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("x-markdown-tokens") == "" {
		t.Fatal("missing x-markdown-tokens")
	}
	if !strings.Contains(rr.Body.String(), "# Info Links") {
		t.Fatalf("body: %s", rr.Body.String())
	}
}

func TestStaticHandler_homepageLinkHeaders(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	links := rr.Header().Values("Link")
	joined := strings.Join(links, ", ")
	for _, want := range []string{
		`rel="api-catalog"`,
		`rel="service-desc"`,
		`rel="service-doc"`,
		`rel="describedby"`,
		`/.well-known/api-catalog`,
		`/openapi.json`,
		`/api/docs`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("Link headers missing %q; got %q", want, joined)
		}
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

func TestStaticHandler_assetCacheControl(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.Mkdir(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(dir, "index.html"):           "<html></html>",
		filepath.Join(assets, "logo.png"):          "png",
		filepath.Join(assets, "index-CQUV9458.js"): "js",
	}
	for path, body := range files {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	handler := newStaticFileHandler(dir, "https://example.com")
	tests := []struct {
		path string
		want string
	}{
		{path: "/assets/logo.png", want: cacheControlStatic},
		{path: "/assets/index-CQUV9458.js", want: cacheControlHashed},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
			}
			if cc := rr.Header().Get("Cache-Control"); cc != tt.want {
				t.Fatalf("Cache-Control: got %q want %q", cc, tt.want)
			}
		})
	}
}
