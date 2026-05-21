package seo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"infolinks-backend/internal/database"
)

func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") != "" {
		database.InitDB()
	}
	m.Run()
}

func TestHandleRobots(t *testing.T) {
	os.Setenv("SITE_BASE_URL", "https://example.com")
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rr := httptest.NewRecorder()
	h.HandleRobots(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Sitemap: https://example.com/sitemap.xml") {
		t.Errorf("robots missing sitemap: %s", body)
	}
	if !strings.Contains(body, "Disallow: /admin") {
		t.Error("robots missing admin disallow")
	}
}

func TestHandleSitemap(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	os.Setenv("SITE_BASE_URL", "https://example.com")
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()
	h.HandleSitemap(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("invalid sitemap xml")
	}
	if !strings.Contains(body, "https://example.com/courses") {
		t.Error("sitemap missing /courses")
	}
}

func TestHandleCourseNotFound(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/course/ZZZZNOTACODE999", nil)
	req.SetPathValue("code", "ZZZZNOTACODE999")
	rr := httptest.NewRecorder()
	h.HandleCourse(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleCourseFound(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	h := NewHandler()
	codes, err := h.Repo.ListCourseCodesForSitemap(context.Background())
	if err != nil || len(codes) == 0 {
		t.Skip("no courses in database")
	}
	code := codes[0]
	req := httptest.NewRequest(http.MethodGet, "/course/"+code, nil)
	req.SetPathValue("code", code)
	rr := httptest.NewRecorder()
	h.HandleCourse(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String()[:200])
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<h1>") {
		t.Error("expected h1 in course page")
	}
	if !strings.Contains(body, "schema.org") {
		t.Error("expected JSON-LD")
	}
}
