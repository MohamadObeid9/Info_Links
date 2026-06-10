package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestShouldSkipMetrics(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/metrics", true},
		{"/healthz", true},
		{"/readyz", true},
		{"/api/content", false},
		{"/api/admin/reports", false},
		{"/main.js", true},
		{"/styles/components.css", true},
	}
	for _, tt := range tests {
		if got := shouldSkipMetrics(tt.path); got != tt.want {
			t.Errorf("shouldSkipMetrics(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/api/content", "/api/content"},
		{"/api/admin/links/42", "/api/admin/links/{id}"},
		{"/api/admin/reports", "/api/admin/reports"},
		{"/course/nfa008", "/course/{code}"},
		{"/program/licence", "/program/{slug}"},
	}
	for _, tt := range tests {
		if got := NormalizePath(tt.path); got != tt.want {
			t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestMetrics_recordsRequest(t *testing.T) {
	handler := Metrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}

	value, ok := counterValue(t, "http_requests_total", map[string]string{
		"method": "GET",
		"path":   "/api/content",
		"status": "200",
	})
	if !ok {
		t.Fatal("http_requests_total{GET,/api/content,200} not found")
	}
	if value < 1 {
		t.Fatalf("expected counter >= 1, got %v", value)
	}
}

func TestMetrics_recordsNonOKStatus(t *testing.T) {
	handler := Metrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d", rr.Code)
	}

	value, ok := counterValue(t, "http_requests_total", map[string]string{
		"method": "GET",
		"path":   "/api/missing",
		"status": "404",
	})
	if !ok {
		t.Fatal("http_requests_total{GET,/api/missing,404} not found")
	}
	if value < 1 {
		t.Fatalf("expected counter >= 1, got %v", value)
	}
}

func TestMetrics_skipsProbePaths(t *testing.T) {
	handler := Metrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(rr, req)
	}

	if hasCounter(t, "http_requests_total", map[string]string{"path": "/healthz"}) {
		t.Fatal("healthz should not be recorded in http_requests_total")
	}
	if hasCounter(t, "http_requests_total", map[string]string{"path": "/readyz"}) {
		t.Fatal("readyz should not be recorded in http_requests_total")
	}
	if hasCounter(t, "http_requests_total", map[string]string{"path": "/metrics"}) {
		t.Fatal("/metrics should not be recorded in http_requests_total")
	}
}

func counterValue(t *testing.T, name string, labels map[string]string) (float64, bool) {
	t.Helper()

	fams, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, fam := range fams {
		if fam.GetName() != name {
			continue
		}
		for _, m := range fam.GetMetric() {
			if !labelsMatch(m, labels) {
				continue
			}
			return m.GetCounter().GetValue(), true
		}
	}
	return 0, false
}

func hasCounter(t *testing.T, name string, labels map[string]string) bool {
	t.Helper()
	_, ok := counterValue(t, name, labels)
	return ok
}

func labelsMatch(m *dto.Metric, want map[string]string) bool {
	got := labelMap(m)
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

func labelMap(m *dto.Metric) map[string]string {
	out := make(map[string]string)
	for _, lp := range m.GetLabel() {
		out[lp.GetName()] = lp.GetValue()
	}
	return out
}
