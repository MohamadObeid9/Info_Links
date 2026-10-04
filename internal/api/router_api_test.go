package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"infolinks-backend/internal/config"
)

func TestRouterAPIIndexAndUnknownPaths(t *testing.T) {
	handler := testRouter(t)

	t.Run("exact /api is the index", func(t *testing.T) {
		body := getJSON(t, handler, http.MethodGet, "/api")
		if _, ok := body["admin_endpoints"]; ok {
			t.Fatal("public index must not list admin_endpoints")
		}
		if body["message"] == nil {
			t.Fatalf("index body = %#v", body)
		}
	})

	t.Run("exact /api/ is the index", func(t *testing.T) {
		body := getJSON(t, handler, http.MethodGet, "/api/")
		if _, ok := body["admin_endpoints"]; ok {
			t.Fatal("public index must not list admin_endpoints")
		}
	})

	t.Run("registered /api/content is not the catch-all", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
		handler.ServeHTTP(rr, req)
		if rr.Code == http.StatusNotFound {
			t.Fatalf("GET /api/content was treated as unknown: %s", rr.Body.String())
		}
	})

	for _, path := range []string{
		"/api/.env",
		"/api/phpinfo.php",
		"/api/content/offerings/999991",
		"/api/does-not-exist",
	} {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			handler.ServeHTTP(rr, req)
			assertNotFoundJSON(t, rr)
		})
	}
}

func TestRouterAdminRoutesRequireAuth(t *testing.T) {
	handler := testRouter(t)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/page_views"},
		{http.MethodGet, "/api/admin/link_clicks"},
		{http.MethodGet, "/api/admin/users"},
		{http.MethodDelete, "/api/admin/users"},
		{http.MethodGet, "/api/admin/users/1"},
		{http.MethodGet, "/api/admin/analytics/summary"},
		{http.MethodGet, "/api/admin/analytics/actors"},
		{http.MethodPost, "/api/admin/links"},
		{http.MethodPatch, "/api/admin/links/1"},
		{http.MethodDelete, "/api/admin/links/1"},
		{http.MethodPost, "/api/admin/courses"},
		{http.MethodPatch, "/api/admin/courses/1"},
		{http.MethodDelete, "/api/admin/courses/1"},
		{http.MethodGet, "/api/admin/reports"},
		{http.MethodPatch, "/api/admin/reports/1"},
		{http.MethodDelete, "/api/admin/reports/1"},
		{http.MethodGet, "/api/admin/feedback"},
		{http.MethodPatch, "/api/admin/feedback/1"},
		{http.MethodDelete, "/api/admin/feedback/1"},
		{http.MethodGet, "/api/admin/contributions"},
		{http.MethodPatch, "/api/admin/contributions/1"},
		{http.MethodDelete, "/api/admin/contributions/1"},
		{http.MethodGet, "/api/admin/extra_sections"},
		{http.MethodPost, "/api/admin/extra_sections"},
		{http.MethodPatch, "/api/admin/extra_sections/1"},
		{http.MethodDelete, "/api/admin/extra_sections/1"},
		{http.MethodGet, "/api/admin/extra_links"},
		{http.MethodPost, "/api/admin/extra_links"},
		{http.MethodPatch, "/api/admin/extra_links/1"},
		{http.MethodDelete, "/api/admin/extra_links/1"},
		{http.MethodGet, "/api/admin/services"},
		{http.MethodPost, "/api/admin/services"},
		{http.MethodPatch, "/api/admin/services/1"},
		{http.MethodDelete, "/api/admin/services/1"},
		{http.MethodPost, "/api/admin/services/1/renew"},
		{http.MethodPost, "/api/admin/services/1/freeze"},
		{http.MethodPost, "/api/admin/services/1/unfreeze"},
	}

	for i, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(rt.method, rt.path, nil)
			req.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", i+1)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 401 or 403, body = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Config{
		AppEnv:    "development",
		JWTSecret: "test-secret",
	}
	return NewRouter(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), testHandler(t), nil)
}

func getJSON(t *testing.T, handler http.Handler, method, path string) map[string]any {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

func assertNotFoundJSON(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body = %s", rr.Code, rr.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if body["error"] != "not found" {
		t.Fatalf("error = %q, want %q", body["error"], "not found")
	}
}
