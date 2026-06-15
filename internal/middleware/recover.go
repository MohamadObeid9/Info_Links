package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover catches panics, logs them, and returns 500 without crashing the process.
func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				id := RequestIDFromContext(r.Context())
				log := logger
				if id != "" {
					log = logger.With("request_id", id)
				}

				log.Error(
					"panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":      "Internal server error",
					"request_id": id,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
