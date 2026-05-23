package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	h.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q want application/json", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("json decode: %v", err)
	}

	want := map[string]string{
		"status":  "ok",
		"message": "The Go backend is alive and healthy!",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response: got %+v want %+v", got, want)
	}
}

func TestHandleApiRoot(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rr := httptest.NewRecorder()

	h.handleApiRoot(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q want application/json", ct)
	}

	var got map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("json decode: %v", err)
	}

	if got["message"] != "Welcome to the Info Links API!" {
		t.Fatalf("message: got %v", got["message"])
	}
	if got["usage"] != "This is a Go backend serving JSON data." {
		t.Fatalf("usage: got %v", got["usage"])
	}

	public, ok := got["public_endpoints"].([]any)
	if !ok {
		t.Fatalf("public_endpoints: got %T", got["public_endpoints"])
	}
	if len(public) != 5 {
		t.Fatalf("public_endpoints len: got %d want 5", len(public))
	}
	assertEndpointCatalog(t, public, []endpointSpec{
		{path: "/api/health", method: "GET", description: "Check if the server is healthy."},
		{path: "/api/content", method: "GET", description: "Fetch the full navigation tree."},
		{path: "/api/auth/login", method: "POST", description: "Admin login (returns JWT token)."},
		{path: "/api/feedback", method: "POST", description: "Submit user feedback."},
		{path: "/api/reports", method: "POST", description: "Submit a course/link report."},
	})

	admin, ok := got["admin_endpoints"].([]any)
	if !ok {
		t.Fatalf("admin_endpoints: got %T", got["admin_endpoints"])
	}
	if len(admin) != 6 {
		t.Fatalf("admin_endpoints len: got %d want 6", len(admin))
	}
	assertEndpointCatalog(t, admin, []endpointSpec{
		{path: "/api/admin/courses", method: "POST/PATCH/DELETE", description: "Manage courses."},
		{path: "/api/admin/links", method: "POST/PATCH/DELETE", description: "Manage links."},
		{path: "/api/admin/reports", method: "GET/PATCH/DELETE", description: "Manage user reports."},
		{path: "/api/admin/feedback", method: "GET/PATCH/DELETE", description: "Manage feedback."},
		{path: "/api/admin/page_views", method: "GET", description: "View analytics (page views)."},
		{path: "/api/admin/link_clicks", method: "GET", description: "View analytics (link clicks)."},
	})
}

type endpointSpec struct {
	path        string
	method      string
	description string
}

func assertEndpointCatalog(t *testing.T, got []any, want []endpointSpec) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("endpoint count: got %d want %d", len(got), len(want))
	}
	for i, spec := range want {
		entry, ok := got[i].(map[string]any)
		if !ok {
			t.Fatalf("endpoint[%d]: got %T", i, got[i])
		}
		if entry["path"] != spec.path {
			t.Fatalf("endpoint[%d].path: got %v want %q", i, entry["path"], spec.path)
		}
		if entry["method"] != spec.method {
			t.Fatalf("endpoint[%d].method: got %v want %q", i, entry["method"], spec.method)
		}
		if entry["description"] != spec.description {
			t.Fatalf("endpoint[%d].description: got %v want %q", i, entry["description"], spec.description)
		}
	}
}
