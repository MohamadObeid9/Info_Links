package api

import (
	"net/http"

	"github.com/rs/cors"
)

// NewRouter sets up the API routes and CORS middleware.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("GET /api", HandleApiRoot)
	mux.HandleFunc("GET /api/", HandleApiRoot)
	mux.HandleFunc("GET /api/health", HandleHealth)
	mux.HandleFunc("GET /api/content", HandleGetContent)

	mux.HandleFunc("POST /api/page_views", HandlePostPageView)
	mux.HandleFunc("POST /api/link_clicks", HandlePostLinkClick)
	mux.HandleFunc("POST /api/reports", HandlePostReport)
	mux.HandleFunc("POST /api/contributions", HandlePostContribution)
	mux.HandleFunc("POST /api/feedback", HandlePostFeedback)

	// Serve the static frontend
	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("/", fs)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	return c.Handler(mux)
}
