package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"infolinks-backend/internal/config"

	"github.com/rs/cors"
)

// NewRouter sets up the API routes and CORS middleware.
func NewRouter(cfg config.Config, h *Handler) http.Handler {
	mux := http.NewServeMux()

	handleAdminFunc := func(pattern string, handler http.HandlerFunc) {
		mux.HandleFunc(pattern, h.requireAdmin(handler))
	}

	// 1. Public API Routes
	mux.HandleFunc("GET /api", h.handleApiRoot)
	mux.HandleFunc("GET /api/", h.handleApiRoot)
	mux.HandleFunc("GET /api/content", h.handleGetContent)
	mux.HandleFunc("GET /api/health", h.handleHealth)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)

	mux.HandleFunc("POST /api/page_views", h.handlePostPageView)
	mux.HandleFunc("POST /api/link_clicks", h.handlePostLinkClick)
	mux.HandleFunc("POST /api/reports", h.handlePostReport)
	mux.HandleFunc("POST /api/contributions", h.handlePostContribution)
	mux.HandleFunc("POST /api/feedback", h.handlePostFeedback)

	// 2. Admin Protected API Routes
	handleAdminFunc("GET /api/admin/reports", h.handleAdminGetReports)
	handleAdminFunc("PATCH /api/admin/reports/{id}", h.handleAdminUpdateReport)
	handleAdminFunc("DELETE /api/admin/reports/{id}", h.handleAdminDeleteReport)

	handleAdminFunc("GET /api/admin/feedback", h.handleAdminGetFeedback)
	handleAdminFunc("PATCH /api/admin/feedback/{id}", h.handleAdminPatchFeedback)
	handleAdminFunc("DELETE /api/admin/feedback/{id}", h.handleAdminDeleteFeedback)

	handleAdminFunc("GET /api/admin/contributions", h.handleAdminGetContributions)
	handleAdminFunc("PATCH /api/admin/contributions/{id}", h.handleAdminUpdateContribution)
	handleAdminFunc("DELETE /api/admin/contributions/{id}", h.handleAdminDeleteContribution)

	handleAdminFunc("POST /api/admin/courses", h.handleAdminPostCourse)
	handleAdminFunc("PATCH /api/admin/courses/{id}", h.handleAdminPatchCourse)
	handleAdminFunc("DELETE /api/admin/courses/{id}", h.handleAdminDeleteCourse)

	handleAdminFunc("POST /api/admin/links", h.handleAdminPostLink)
	handleAdminFunc("PATCH /api/admin/links/{id}", h.handleAdminPatchLink)
	handleAdminFunc("DELETE /api/admin/links/{id}", h.handleAdminDeleteLink)

	handleAdminFunc("GET /api/admin/page_views", h.handleAdminGetPageViews)
	handleAdminFunc("GET /api/admin/link_clicks", h.handleAdminGetLinkClicks)

	// 3. Static Files & SPA Routing
	staticDir := "frontend/dist"
	if _, err := os.Stat(staticDir); err != nil {
		staticDir = "frontend"
	}

	fs := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the file exists, serve it
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

	var parsedOrigins []string
	allowedOrigins := []string{"http://localhost:8080", "http://localhost:5173"}
	if raw := strings.TrimSpace(cfg.CorsAllowedOrigins); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			origin := strings.TrimSpace(item)
			if origin != "" {
				parsedOrigins = append(parsedOrigins, origin)
			}
		}
	}
	if len(parsedOrigins) > 0 {
		allowedOrigins = parsedOrigins
	}

	connectSrcValues := []string{"'self'"}
	for _, origin := range allowedOrigins {
		if origin != "" {
			connectSrcValues = append(connectSrcValues, origin)
		}
	}
	connectSrc := strings.Join(connectSrcValues, " ")
	cspValue := strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' ",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"img-src 'self' data:",
		"font-src 'self' https://fonts.gstatic.com",
		"connect-src " + connectSrc,
		"object-src 'none'",
		"base-uri 'self'",
		"frame-ancestors 'none'",
	}, "; ")

	// 4. CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
	})
	withSecurityHeaders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers are applied centrally to both API and static responses.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", cspValue)
		mux.ServeHTTP(w, r)
	})

	return c.Handler(withSecurityHeaders)
}
