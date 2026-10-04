package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestShouldSkipMetrics(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"GET /metrics", true},
		{"GET /healthz", true},
		{"GET /readyz", true},
		{"GET /api/content", false},
		{"GET /api/admin/reports", false},
		{"/", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := shouldSkipMetrics(tt.pattern); got != tt.want {
			t.Errorf("shouldSkipMetrics(%q) = %v, want %v", tt.pattern, got, tt.want)
		}
	}
}

func TestMetrics_recordsMatchedPattern(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/content", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /api/users/me/favorites/{course_id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := Metrics(mux)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/content", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if _, ok := counterValue(t, "http_requests_total", map[string]string{
		"method": "GET",
		"path":   "/api/content",
		"status": "200",
	}); !ok {
		t.Fatal("http_requests_total{GET,/api/content,200} not found")
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/users/me/favorites/22", nil))
	if _, ok := counterValue(t, "http_requests_total", map[string]string{
		"method": "POST",
		"path":   "/api/users/me/favorites/{course_id}",
		"status": "201",
	}); !ok {
		t.Fatal("favorite route was not labeled with its pattern")
	}
}

func TestMetrics_otherMethod(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	handler := Metrics(mux)

	before := pathCounts(t, "http_requests_total")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("TRACE", "/api/.env", nil))

	after := pathCounts(t, "http_requests_total")
	if after["unmatched"]-before["unmatched"] != 1 {
		t.Fatalf("unmatched delta = %v, want 1", after["unmatched"]-before["unmatched"])
	}
	if _, ok := counterValue(t, "http_requests_total", map[string]string{
		"method": "other",
		"path":   metricsUnmatchedPath,
		"status": "404",
	}); !ok {
		t.Fatal("TRACE was not labeled method=other path=unmatched")
	}
}

func TestMetrics_randomPathsCollapseToUnmatched(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	handler := Metrics(mux)

	before := pathCounts(t, "http_requests_total")
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 50; i++ {
		path := fmt.Sprintf("/scan/%d/%d", i, rng.Int())
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	changed := map[string]float64{}
	after := pathCounts(t, "http_requests_total")
	for path, n := range after {
		if d := n - before[path]; d > 0 {
			changed[path] = d
		}
	}
	if len(changed) != 1 || changed[metricsUnmatchedPath] != 50 {
		t.Fatalf("distinct path labels = %v, want exactly {unmatched: 50}", changed)
	}
}

func TestMetrics_skipsProbePatterns(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {})
	handler := Metrics(mux)

	before := pathCounts(t, "http_requests_total")
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	after := pathCounts(t, "http_requests_total")
	for path, n := range after {
		if n > before[path] {
			t.Fatalf("probe request recorded path %q", path)
		}
	}
}

func pathCounts(t *testing.T, name string) map[string]float64 {
	t.Helper()
	out := map[string]float64{}
	fams, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, fam := range fams {
		if fam.GetName() != name {
			continue
		}
		for _, m := range fam.GetMetric() {
			labels := labelMap(m)
			out[labels["path"]] += m.GetCounter().GetValue()
		}
	}
	return out
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
