package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestHandleHealthz(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.handleHealthz(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q want application/json", ct)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != `"ok"` {
		t.Fatalf("body: got %q want %q", body, `"ok"`)
	}
}

func TestHandleReadyz(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string // exact body for success; error message for failure
	}{
		{
			name:       "200 when database ping succeeds",
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantBody:   `"ready"`,
		},
		{
			name:       "503 when database ping fails",
			pingErr:    errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "db is unreachable",
		},
		{
			name:       "503 when database client returns timeout",
			pingErr:    context.DeadlineExceeded,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "db is unreachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t, withDB(&mockPinger{err: tt.pingErr}))
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rr := httptest.NewRecorder()

			h.handleReadyz(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type: got %q want application/json", ct)
			}

			if tt.wantStatus == http.StatusOK {
				if body := strings.TrimSpace(rr.Body.String()); body != tt.wantBody {
					t.Fatalf("body: got %q want %q", body, tt.wantBody)
				}
				return
			}

			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["error"] != tt.wantBody {
				t.Fatalf("error: got %q want %q", got["error"], tt.wantBody)
			}
		})
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

	want := map[string]any{
		"message": "Welcome to the Info Links API!",
		"usage":   "This is a Go backend serving JSON data.",
		"public_endpoints": []any{
			map[string]any{"path": "/api/content", "method": "GET", "description": "Fetch the full navigation tree."},
			map[string]any{"path": "/api/auth/login", "method": "POST", "description": "Admin login (returns JWT token)."},
			map[string]any{"path": "/api/feedback", "method": "POST", "description": "Submit user feedback."},
			map[string]any{"path": "/api/reports", "method": "POST", "description": "Submit a course/link report."},
		},
		"admin_endpoints": []any{
			map[string]any{"path": "/api/admin/courses", "method": "POST/PATCH/DELETE", "description": "Manage courses."},
			map[string]any{"path": "/api/admin/links", "method": "POST/PATCH/DELETE", "description": "Manage links."},
			map[string]any{"path": "/api/admin/reports", "method": "GET/PATCH/DELETE", "description": "Manage user reports."},
			map[string]any{"path": "/api/admin/feedback", "method": "GET/PATCH/DELETE", "description": "Manage feedback."},
			map[string]any{"path": "/api/admin/page_views", "method": "GET", "description": "View analytics (page views)."},
			map[string]any{"path": "/api/admin/link_clicks", "method": "GET", "description": "View analytics (link clicks)."},
		},
	}

	lessFunc := func(a, b map[string]any) bool {
		return fmt.Sprintf("%v", a["path"]) < fmt.Sprintf("%v", b["path"])
	}

	if diff := cmp.Diff(want, got, cmpopts.SortSlices(lessFunc)); diff != "" {
		t.Fatalf("handleApiRoot response mismatch (-want +got):\n%s", diff)
	}

}
