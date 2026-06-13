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
	registerAdminRoutes(mux, h)
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

func registerAdminRoutes(mux *http.ServeMux, h *Handler) {
	handleAdminFunc := func(pattern string, handler http.HandlerFunc) {
		mux.HandleFunc(pattern, h.requireAdmin(handler))
	}
	handleAdminFunc("GET /api/admin/page_views", h.handleAdminGetPageViews)
	handleAdminFunc("GET /api/admin/link_clicks", h.handleAdminGetLinkClicks)

	handleAdminFunc("POST /api/admin/links", h.handleAdminPostLink)
	handleAdminFunc("PATCH /api/admin/links/{id}", h.handleAdminPatchLink)
	handleAdminFunc("DELETE /api/admin/links/{id}", h.handleAdminDeleteLink)

	handleAdminFunc("POST /api/admin/courses", h.handleAdminPostCourse)
	handleAdminFunc("PATCH /api/admin/courses/{id}", h.handleAdminPatchCourse)
	handleAdminFunc("DELETE /api/admin/courses/{id}", h.handleAdminDeleteCourse)

	handleAdminFunc("GET /api/admin/reports", h.handleAdminGetReports)
	handleAdminFunc("PATCH /api/admin/reports/{id}", h.handleAdminUpdateReport)
	handleAdminFunc("DELETE /api/admin/reports/{id}", h.handleAdminDeleteReport)

	handleAdminFunc("GET /api/admin/feedback", h.handleAdminGetFeedback)
	handleAdminFunc("PATCH /api/admin/feedback/{id}", h.handleAdminPatchFeedback)
	handleAdminFunc("DELETE /api/admin/feedback/{id}", h.handleAdminDeleteFeedback)

	handleAdminFunc("GET /api/admin/contributions", h.handleAdminGetContributions)
	handleAdminFunc("PATCH /api/admin/contributions/{id}", h.handleAdminUpdateContribution)
	handleAdminFunc("DELETE /api/admin/contributions/{id}", h.handleAdminDeleteContribution)

	handleAdminFunc("GET /api/admin/extra_sections", h.handleAdminGetExtraSections)
	handleAdminFunc("POST /api/admin/extra_sections", h.handleAdminPostExtraSection)
	handleAdminFunc("PATCH /api/admin/extra_sections/{id}", h.handleAdminPatchExtraSection)
	handleAdminFunc("DELETE /api/admin/extra_sections/{id}", h.handleAdminDeleteExtraSection)

	handleAdminFunc("GET /api/admin/extra_links", h.handleAdminGetExtraLinks)
	handleAdminFunc("POST /api/admin/extra_links", h.handleAdminPostExtraLink)
	handleAdminFunc("PATCH /api/admin/extra_links/{id}", h.handleAdminPatchExtraLink)
	handleAdminFunc("DELETE /api/admin/extra_links/{id}", h.handleAdminDeleteExtraLink)
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
				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if info.IsDir() && r.URL.Path != "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
	})
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
