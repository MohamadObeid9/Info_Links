package middleware

import (
	"net/http"
	"testing"
)

func Test_accessLogDecision(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		appEnv string
		status int
		want   accessLogAction
	}{
		{
			name: "skip js files",
			path: "/main.js",
			want: accessLogSkip,
		},
		{
			name: "skip html files",
			path: "/main.html",
			want: accessLogSkip,
		},
		{
			name: "skip css files",
			path: "/main.css",
			want: accessLogSkip,
		},
		{
			name: "skip assets",
			path: "/assets/photo.png",
			want: accessLogSkip,
		},
		{
			name: "skip  health check /readyz",
			path: "/readyz",
			want: accessLogSkip,
		},
		{
			name: "skip  health check /healthz",
			path: "/healthz",
			want: accessLogSkip,
		},
		{
			name: "skip /robots.txt path",
			path: "/robots.txt",
			want: accessLogSkip,
		},
		{
			name: "skip /sitemap.xm path",
			path: "/sitemap.xml",
			want: accessLogSkip,
		},
		{
			name: "skip prometheus hitting /metrics",
			path: "/metrics",
			want: accessLogSkip,
		},
		{
			name: "skip php paths",
			path: "/wp-login.php",
			want: accessLogSkip,
		},
		{
			name: "skip env paths",
			path: "/app/.env",
			want: accessLogSkip,
		},
		{
			name: "skip unidentified files",
			path: "/something.strange",
			want: accessLogSkip,
		},
		{
			name:   "skip HEAD methods",
			method: http.MethodHead,
			want:   accessLogSkip,
		},
		{
			name: "debug /api/admin paths",
			path: "/api/admin/reports",
			want: accessLogSkip,
		},
		{
			name:   "info rate limit",
			status: http.StatusTooManyRequests,
			want:   accessLogWarn,
		},
		{
			name:   "info delete method",
			method: http.MethodDelete,
			want:   accessLogInfo,
		},
		{
			name:   "info patch method",
			method: http.MethodPatch,
			want:   accessLogInfo,
		},
		{
			name:   "info post method",
			method: http.MethodPost,
			want:   accessLogInfo,
		},
		{
			name:   "warn status = 400",
			status: http.StatusBadRequest,
			want:   accessLogWarn,
		},
		{
			name:   "warn status > 400",
			status: http.StatusUnauthorized,
			want:   accessLogWarn,
		},
		{
			name:   "warn status not found",
			status: http.StatusNotFound,
			want:   accessLogWarn,
		},
		{
			name:   "info default if appEnv = development",
			appEnv: "development",
			want:   accessLogInfo,
		},
		{
			name:   "debug default if appEnv = production",
			appEnv: "production",
			want:   accessLogDebug,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessLogDecision(tt.method, tt.path, tt.appEnv, tt.status)
			if got != tt.want {
				t.Fatalf("got %v , want %v", got, tt.want)
			}
		})
	}
}

func Test_isNoisyPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
		want   bool
	}{
		{
			name: "skip js files",
			path: "/main.js",
			want: true,
		},
		{
			name: "skip html files",
			path: "/main.html",
			want: true,
		},
		{
			name: "skip css files",
			path: "/main.css",
			want: true,
		},
		{
			name: "skip assets",
			path: "/assets/photo.png",
			want: true,
		},
		{
			name: "skip  health check /readyz",
			path: "/readyz",
			want: true,
		},
		{
			name: "skip  health check /healthz",
			path: "/healthz",
			want: true,
		},
		{
			name: "skip /robots.txt path",
			path: "/robots.txt",
			want: true,
		},
		{
			name: "skip /sitemap.xm path",
			path: "/sitemap.xml",
			want: true,
		},
		{
			name: "skip prometheus hitting /metrics",
			path: "/metrics",
			want: true,
		},
		{
			name: "skip php paths",
			path: "/wp-login.php",
			want: true,
		},
		{
			name: "skip env paths",
			path: "/app/.env",
			want: true,
		},
		{
			name: "skip unidentified files",
			path: "/something.strange",
			want: true,
		},
		{
			name:   "skip HEAD methods",
			method: http.MethodHead,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNoisyPath(tt.path, tt.method)
			if got != tt.want {
				t.Fatalf("got %v , want %v", got, tt.want)
			}
		})
	}
}
