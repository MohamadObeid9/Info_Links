package middleware

import (
	"crypto/subtle"
	"net/http"
)

const metricsAuthRealm = `Basic realm="metrics"`

// MetricsBasicAuth protects an HTTP handler with HTTP Basic authentication.
func MetricsBasicAuth(user, pass string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(username), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(password), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", metricsAuthRealm)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MetricsDenied rejects all requests to the metrics endpoint.
func MetricsDenied() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", metricsAuthRealm)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
