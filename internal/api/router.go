package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/cors"
)

// NewRouter sets up the API routes and CORS middleware.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// 1. Public API Routes
	mux.HandleFunc("GET /api", HandleApiRoot)
	mux.HandleFunc("GET /api/", HandleApiRoot)
	mux.HandleFunc("GET /api/content", HandleGetContent)
	mux.HandleFunc("GET /api/health", HandleHealth)
	mux.HandleFunc("POST /api/auth/login", HandleLogin)

	mux.HandleFunc("POST /api/page_views", HandlePostPageView)
	mux.HandleFunc("POST /api/link_clicks", HandlePostLinkClick)
	mux.HandleFunc("POST /api/reports", HandlePostReport)
	mux.HandleFunc("POST /api/contributions", HandlePostContribution)
	mux.HandleFunc("POST /api/feedback", HandlePostFeedback)

	// 2. Admin Protected API Routes
	mux.HandleFunc("GET /api/admin/reports", RequireAdmin(HandleAdminGetReports))
	mux.HandleFunc("PATCH /api/admin/reports/{id}", RequireAdmin(HandleAdminUpdateReport))
	mux.HandleFunc("DELETE /api/admin/reports/{id}", RequireAdmin(HandleAdminDeleteReport))

	mux.HandleFunc("GET /api/admin/feedback", RequireAdmin(HandleAdminGetFeedback))
	mux.HandleFunc("PATCH /api/admin/feedback/{id}", RequireAdmin(HandleAdminPatchFeedback))
	mux.HandleFunc("DELETE /api/admin/feedback/{id}", RequireAdmin(HandleAdminDeleteFeedback))

	mux.HandleFunc("GET /api/admin/contributions", RequireAdmin(HandleAdminGetContributions))
	mux.HandleFunc("PATCH /api/admin/contributions/{id}", RequireAdmin(HandleAdminUpdateContribution))
	mux.HandleFunc("DELETE /api/admin/contributions/{id}", RequireAdmin(HandleAdminDeleteContribution))

	mux.HandleFunc("POST /api/admin/courses", RequireAdmin(HandleAdminPostCourse))
	mux.HandleFunc("PATCH /api/admin/courses/{id}", RequireAdmin(HandleAdminPatchCourse))
	mux.HandleFunc("DELETE /api/admin/courses/{id}", RequireAdmin(HandleAdminDeleteCourse))

	mux.HandleFunc("POST /api/admin/links", RequireAdmin(HandleAdminPostLink))
	mux.HandleFunc("PATCH /api/admin/links/{id}", RequireAdmin(HandleAdminPatchLink))
	mux.HandleFunc("DELETE /api/admin/links/{id}", RequireAdmin(HandleAdminDeleteLink))

	mux.HandleFunc("GET /api/admin/page_views", RequireAdmin(HandleAdminGetPageViews))
	mux.HandleFunc("GET /api/admin/link_clicks", RequireAdmin(HandleAdminGetLinkClicks))

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
		if os.IsNotExist(err) || info.IsDir() {
			// If it doesn't exist or is a directory (and we want index.html), serve index.html
			// We check if it's a directory to avoid serving directory listings
			if info != nil && info.IsDir() && r.URL.Path != "/" {
				// If it's a sub-directory, we still serve index.html for the SPA
				http.ServeFile(w, r, staticDir+"/index.html")
				return
			}
			if os.IsNotExist(err) {
				http.ServeFile(w, r, staticDir+"/index.html")
				return
			}
		}
		// Otherwise serve the file
		fs.ServeHTTP(w, r)
	})

	allowedOrigins := []string{"http://localhost:8080", "http://localhost:5173"}
	if raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); raw != "" {
		allowedOrigins = nil
		for _, item := range strings.Split(raw, ",") {
			origin := strings.TrimSpace(item)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
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
		"script-src 'self' 'unsafe-inline'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"img-src 'self' data:",
		"font-src 'self' https://fonts.gstatic.com",
		"connect-src " + connectSrc,
		"object-src 'none'",
		"base-uri 'self'",
		"frame-ancestors 'none'",
	}, "; ")

	// 3. CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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
