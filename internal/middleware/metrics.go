package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
		if shouldSkipMetrics(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		path := NormalizePath(r.URL.Path)
		status := strconv.Itoa(ww.status)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestsDuration.WithLabelValues(r.Method, path).Observe(float64(time.Since(start).Seconds()))
	})
}

func shouldSkipMetrics(path string) bool {
	switch path {
	case "/metrics", "/healthz", "/readyz":
		return true
	}
	if strings.HasPrefix(path, "/assets/") {
		return true
	}
	if dot := strings.LastIndex(path, "."); dot != -1 {
		ext := path[dot:]
		switch ext {
		case ".js", ".css", ".ico", ".png", ".webp", ".svg", ".woff2", ".map":
			return true
		}
	}
	return false
}

func NormalizePath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return path
	}
	switch {
	case parts[0] == "api" && len(parts) >= 4 && parts[1] == "admin":
		if parts[3] != "" {
			parts[3] = "{id}"
			return "/" + strings.Join(parts, "/")
		}
	case parts[0] == "course" && len(parts) >= 2:
		return "/course/{code}"
	case parts[0] == "program" && len(parts) >= 2:
		return "/program/{slug}"
	}
	return path
}
