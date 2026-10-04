package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsUnmatchedPath = "unmatched"

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		if shouldSkipMetrics(r.Pattern) {
			return
		}

		method := metricMethod(r.Method)
		path := metricPath(r.Pattern)
		status := strconv.Itoa(ww.status)

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestsDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	})
}

func shouldSkipMetrics(pattern string) bool {
	switch metricPath(pattern) {
	case "/metrics", "/healthz", "/readyz":
		return true
	default:
		return false
	}
}

// metricPath is the ServeMux pattern with the method prefix removed.
// An empty match and the SPA catch-all "/" share one label so raw URLs never become series.
func metricPath(pattern string) string {
	path := pattern
	if i := strings.IndexByte(pattern, ' '); i >= 0 {
		path = pattern[i+1:]
	}
	if path == "" || path == "/" {
		return metricsUnmatchedPath
	}
	return path
}

func metricMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodHead:
		return method
	default:
		return "other"
	}
}
