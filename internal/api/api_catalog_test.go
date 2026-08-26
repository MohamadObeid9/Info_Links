package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAPICatalog(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/api-catalog", nil)
	rr := httptest.NewRecorder()

	h.handleAPICatalog(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200 body=%q", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/linkset+json") {
		t.Fatalf("Content-Type: got %q", ct)
	}
	if !strings.Contains(rr.Header().Get("Link"), `rel="api-catalog"`) {
		t.Fatalf("Link header: got %q", rr.Header().Get("Link"))
	}

	var doc apiCatalogDocument
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v body=%s", err, rr.Body.String())
	}
	if len(doc.Linkset) == 0 {
		t.Fatal("empty linkset")
	}
	entry := doc.Linkset[0]
	if entry.Anchor != "https://example.com/api" {
		t.Fatalf("anchor: got %q", entry.Anchor)
	}
	if len(entry.ServiceDesc) == 0 || entry.ServiceDesc[0].Href != "https://example.com/openapi.json" {
		t.Fatalf("service-desc: %+v", entry.ServiceDesc)
	}
	if len(entry.ServiceDoc) == 0 || entry.ServiceDoc[0].Href != "https://example.com/api/docs" {
		t.Fatalf("service-doc: %+v", entry.ServiceDoc)
	}
	if len(entry.Status) == 0 || entry.Status[0].Href != "https://example.com/healthz" {
		t.Fatalf("status: %+v", entry.Status)
	}
}

func TestHandleAPICatalogHEAD(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodHead, "/.well-known/api-catalog", nil)
	rr := httptest.NewRecorder()

	h.handleAPICatalog(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("HEAD body should be empty, got %q", rr.Body.String())
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/linkset+json") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
}

func TestHandleOpenAPI(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rr := httptest.NewRecorder()

	h.handleOpenAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "openapi+json") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
	var spec map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("openapi json: %v", err)
	}
	if spec["openapi"] == nil {
		t.Fatal("missing openapi version field")
	}
}

func TestHandleAPIDocs(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rr := httptest.NewRecorder()

	h.handleAPIDocs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
	if !strings.Contains(rr.Body.String(), "# Info Links API") {
		t.Fatalf("body: %s", rr.Body.String())
	}
}

func TestRouterAPICatalog(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/api-catalog", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/linkset+json") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
}
