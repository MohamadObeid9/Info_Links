package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const HeaderRequestID = "X-Request-ID"

type contextKey struct{}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isValidRequestID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' {
			continue
		}
		return false
	}
	return true
}

// RequestIDFromContext returns the request ID stored by the middleware.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestID ensures every request has an X-Request-ID without access logging.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestIDFromRequest(r)
		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), contextKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDWithLogging wraps RequestID and logs one access line per request.
func RequestIDWithLogging(logger *slog.Logger, appEnv string, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestIDFromRequest(r)
		w.Header().Set(HeaderRequestID, id)

		ctx := context.WithValue(r.Context(), contextKey{}, id)
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r.WithContext(ctx))

		path := r.URL.Path
		if !strings.Contains(path, "js") && !strings.Contains(path, "styles") && !strings.Contains(path, "assets") && !strings.Contains(path, "html") {
			if (appEnv == "production" && path != "/metrics") || appEnv == "development" {
				if path != "/readyz" && r.Method != "HEAD" {
					logger.Info("request completed",
						"request_id", id,
						"method", r.Method,
						"path", path,
						"status", ww.status,
						"duration_ms", time.Since(start).Milliseconds(),
					)
				}
			}
		}
	})
}

func requestIDFromRequest(r *http.Request) string {
	id := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if !isValidRequestID(id) {
		return newRequestID()
	}
	return id
}
