package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRouterRobotsTxt(t *testing.T) {
	os.Setenv("SITE_BASE_URL", "http://localhost:8080")
	handler := NewRouter()

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
	os.Setenv("SITE_BASE_URL", "http://localhost:8080")
	handler := NewRouter()

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
