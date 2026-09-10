package middleware

import (
	"crypto/subtle"
	"net/http"
)

const cloudflareSecretHeader = "X-Cloudflare-Secret"

// RequireCloudflareSecret blocks requests that bypass Cloudflare when a shared
// secret is configured. Requests must present a matching X-Cloudflare-Secret
// header. An empty secret disables enforcement (local development).
//
// Origin-direct operational paths (/healthz, /readyz, /metrics) are exempt so
// Render probes and Grafana scrapes keep working.
func RequireCloudflareSecret(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" || isCloudflareSecretExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		got := r.Header.Get(cloudflareSecretHeader)
		if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			http.Error(w, "Forbidden - Direct Origin Access Blocked", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isCloudflareSecretExempt(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/metrics":
		return true
	default:
		return false
	}
}
