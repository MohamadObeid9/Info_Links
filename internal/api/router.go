package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"infolinks-backend/internal/config"
	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/seo"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func NewRouter(cfg config.Config, logger *slog.Logger, h *Handler, seoH *seo.Handler) http.Handler {
	mux := http.NewServeMux()

	registerPublicRoutes(mux, h, cfg)
	registerAdminRoutes(cfg.JWTSecret, mux, h)
	registerSEORoutes(mux, seoH)
	mux.Handle("/", newStaticFileHandler(resolveStaticDir()))

	origins := allowedOrigins(cfg.CorsAllowedOrigins)
	securedHandler := withSecurityHeaders(mux, contentSecurityPolicy(origins))
	handlerWithRecover := middleware.Recover(logger, securedHandler)
	handlerWithRateLimit := middleware.RateLimit(handlerWithRecover)
	handlerWithMetrics := middleware.Metrics(handlerWithRateLimit)
	handlerWithRequestID := middleware.RequestIDWithLogging(logger, cfg.AppEnv, handlerWithMetrics)

	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
	}).Handler(handlerWithRequestID)
}

func registerPublicRoutes(mux *http.ServeMux, h *Handler, cfg config.Config) {
	mux.Handle("GET /metrics", metricsHandler(cfg))

	mux.HandleFunc("GET /api", h.handleApiRoot)
	mux.HandleFunc("GET /api/", h.handleApiRoot)
	mux.HandleFunc("GET /readyz", h.handleReadyz)
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("GET /api/content", h.handleGetContent)

	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/reports", h.handlePostReport)
	mux.HandleFunc("POST /api/feedback", h.handlePostFeedback)
	mux.HandleFunc("POST /api/page_views", h.handlePostPageView)
	mux.HandleFunc("POST /api/link_clicks", h.handlePostLinkClick)
	mux.HandleFunc("POST /api/contributions", h.handlePostContribution)
}

func metricsHandler(cfg config.Config) http.Handler {
	handler := promhttp.Handler()
	switch {
	case cfg.MetricsAuthEnabled():
		return middleware.MetricsBasicAuth(
			cfg.MetricsBasicAuthUser,
			cfg.MetricsBasicAuthPass,
			handler,
		)
	case cfg.AppEnv == "production":
		return middleware.MetricsDenied()
	default:
		return handler
	}
}

func registerAdminRoutes(jwtSecret string, mux *http.ServeMux, h *Handler) {

	mux.HandleFunc("GET /api/admin/page_views", middleware.RequireAdmin(jwtSecret, h.handleAdminGetPageViews))
	mux.HandleFunc("GET /api/admin/link_clicks", middleware.RequireAdmin(jwtSecret, h.handleAdminGetLinkClicks))

	mux.HandleFunc("POST /api/admin/links", middleware.RequireAdmin(jwtSecret, h.handleAdminPostLink))
	mux.HandleFunc("PATCH /api/admin/links/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminPatchLink))
	mux.HandleFunc("DELETE /api/admin/links/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteLink))

	mux.HandleFunc("POST /api/admin/courses", middleware.RequireAdmin(jwtSecret, h.handleAdminPostCourse))
	mux.HandleFunc("PATCH /api/admin/courses/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminPatchCourse))
	mux.HandleFunc("DELETE /api/admin/courses/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteCourse))

	mux.HandleFunc("GET /api/admin/reports", middleware.RequireAdmin(jwtSecret, h.handleAdminGetReports))
	mux.HandleFunc("PATCH /api/admin/reports/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminUpdateReport))
	mux.HandleFunc("DELETE /api/admin/reports/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteReport))

	mux.HandleFunc("GET /api/admin/feedback", middleware.RequireAdmin(jwtSecret, h.handleAdminGetFeedback))
	mux.HandleFunc("PATCH /api/admin/feedback/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminPatchFeedback))
	mux.HandleFunc("DELETE /api/admin/feedback/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteFeedback))

	mux.HandleFunc("GET /api/admin/contributions", middleware.RequireAdmin(jwtSecret, h.handleAdminGetContributions))
	mux.HandleFunc("PATCH /api/admin/contributions/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminUpdateContribution))
	mux.HandleFunc("DELETE /api/admin/contributions/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteContribution))

	mux.HandleFunc("GET /api/admin/extra_sections", middleware.RequireAdmin(jwtSecret, h.handleAdminGetExtraSections))
	mux.HandleFunc("POST /api/admin/extra_sections", middleware.RequireAdmin(jwtSecret, h.handleAdminPostExtraSection))
	mux.HandleFunc("PATCH /api/admin/extra_sections/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminPatchExtraSection))
	mux.HandleFunc("DELETE /api/admin/extra_sections/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteExtraSection))

	mux.HandleFunc("GET /api/admin/extra_links", middleware.RequireAdmin(jwtSecret, h.handleAdminGetExtraLinks))
	mux.HandleFunc("POST /api/admin/extra_links", middleware.RequireAdmin(jwtSecret, h.handleAdminPostExtraLink))
	mux.HandleFunc("PATCH /api/admin/extra_links/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminPatchExtraLink))
	mux.HandleFunc("DELETE /api/admin/extra_links/{id}", middleware.RequireAdmin(jwtSecret, h.handleAdminDeleteExtraLink))
}

func registerSEORoutes(mux *http.ServeMux, seoH *seo.Handler) {
	if seoH == nil {
		return
	}
	mux.HandleFunc("/course/{code}", seoH.HandleCourse)
	mux.HandleFunc("/program/{slug}", seoH.HandleProgram)
	mux.HandleFunc("/courses", seoH.HandleCoursesIndex)
	mux.HandleFunc("/sitemap.xml", seoH.HandleSitemap)
	mux.HandleFunc("/robots.txt", seoH.HandleRobots)
}

func resolveStaticDir() string {
	staticDir := "frontend/dist"
	if _, err := os.Stat(staticDir); err != nil {
		return "frontend"
	}
	return staticDir
}

func newStaticFileHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSEOPath(r.URL.Path) {
			http.NotFound(w, r)
			return
		}

		relPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		path := filepath.Join(staticDir, relPath)

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if isSPAPath(r.URL.Path) {
					http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
					return
				}
				http.NotFound(w, r)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if info.IsDir() && r.URL.Path != "/" {
			if isSPAPath(r.URL.Path) {
				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
				return
			}
			http.NotFound(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	})
}

func isSPAPath(path string) bool {
	switch path {
	case "/", "/report-submit", "/feedback", "/admin-gate", "/admin":
		return true
	default:
		return false
	}
}

func isSEOPath(path string) bool {
	return strings.HasPrefix(path, "/course/") ||
		strings.HasPrefix(path, "/program/") ||
		path == "/courses" ||
		path == "/sitemap.xml" ||
		path == "/robots.txt"
}

func allowedOrigins(rawOrigins string) []string {
	defaultOrigins := []string{"http://localhost:8080", "http://localhost:5173"}

	rawOrigins = strings.TrimSpace(rawOrigins)
	if rawOrigins == "" {
		return defaultOrigins
	}

	var origins []string
	for _, item := range strings.Split(rawOrigins, ",") {
		origin := strings.TrimSpace(item)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	if len(origins) == 0 {
		return defaultOrigins
	}

	return origins
}

func contentSecurityPolicy(allowedOrigins []string) string {
	connectSrcValues := []string{"'self'"}
	for _, origin := range allowedOrigins {
		if origin != "" {
			connectSrcValues = append(connectSrcValues, origin)
		}
	}

	return strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"img-src 'self' data:",
		"font-src 'self' https://fonts.gstatic.com",
		"connect-src " + strings.Join(connectSrcValues, " "),
		"object-src 'none'",
		"base-uri 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

func withSecurityHeaders(next http.Handler, cspValue string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", cspValue)

		next.ServeHTTP(w, r)
	})
}
