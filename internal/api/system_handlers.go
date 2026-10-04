package api

import (
	"context"
	"net/http"
	"time"
)

/*  handleHealthz checks if the site is healthy
 *  unused in our case
 *  handleHealthz checks if the site is healthy
 */
func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, "ok")
}

// handleReadyz checks if the site is ready , used on render
func (h *Handler) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.LoggerWithID(r).Warn("readiness check failed", "error", err)
		writeJSONError(w, r, http.StatusServiceUnavailable, "db is unreachable")
		return
	}

	writeJSON(w, http.StatusOK, "ready")
}

// HandleApiRoot provides a simple directory of available endpoints.
func (h *Handler) handleApiRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"message": "Welcome to the Info Links API!",
		"usage":   "This is a Go backend serving JSON data.",
		"public_endpoints": []map[string]string{
			{"path": "/api/content", "method": "GET", "description": "Fetch the full navigation tree."},
			{"path": "/api/auth/login", "method": "POST", "description": "Admin login (returns JWT token)."},
			{"path": "/api/feedback", "method": "POST", "description": "Submit user feedback."},
			{"path": "/api/reports", "method": "POST", "description": "Submit a course/link report."},
			{"path": "/api/page_views", "method": "POST", "description": "Record a page view (analytics)."},
			{"path": "/api/link_clicks", "method": "POST", "description": "Record a link click (analytics)."},
			{"path": "/api/search_events", "method": "POST", "description": "Record a search query (analytics)."},
			{"path": "/api/contributions", "method": "POST", "description": "Submit a user contribution."},
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func handleAPINotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}
