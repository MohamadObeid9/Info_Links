package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsBasicAuth_allowsValidCredentials(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := MetricsBasicAuth("scraper", "secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("scraper", "secret")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestMetricsBasicAuth_rejectsMissingCredentials(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := MetricsBasicAuth("scraper", "secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
	if got := rr.Header().Get("WWW-Authenticate"); got != metricsAuthRealm {
		t.Fatalf("WWW-Authenticate: got %q want %q", got, metricsAuthRealm)
	}
}

func TestMetricsBasicAuth_rejectsInvalidCredentials(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := MetricsBasicAuth("scraper", "secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("scraper", "wrong")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestMetricsDenied_rejectsAllRequests(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	MetricsDenied().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}
