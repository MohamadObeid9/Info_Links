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
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.logger.Warn("readiness check failed", "error", err)
		writeJSONError(w, http.StatusServiceUnavailable, "db is unreachable")
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
		},
		"admin_endpoints": []map[string]string{
			{"path": "/api/admin/courses", "method": "POST/PATCH/DELETE", "description": "Manage courses."},
			{"path": "/api/admin/links", "method": "POST/PATCH/DELETE", "description": "Manage links."},
			{"path": "/api/admin/reports", "method": "GET/PATCH/DELETE", "description": "Manage user reports."},
			{"path": "/api/admin/feedback", "method": "GET/PATCH/DELETE", "description": "Manage feedback."},
			{"path": "/api/admin/page_views", "method": "GET", "description": "View analytics (page views)."},
			{"path": "/api/admin/link_clicks", "method": "GET", "description": "View analytics (link clicks)."},
		},
	}

	writeJSON(w, http.StatusOK, response)
}
