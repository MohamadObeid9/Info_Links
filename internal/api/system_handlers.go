package api

import "net/http"

// HandleHealth checks if the server is healthy.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"message": "The Go backend is alive and healthy!",
	}
	writeJSON(w, http.StatusOK, response)
}

// HandleApiRoot provides a simple directory of available endpoints.
func (h *Handler) handleApiRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"message": "Welcome to the Info Links API!",
		"usage":   "This is a Go backend serving JSON data.",
		"public_endpoints": []map[string]string{
			{"path": "/api/health", "method": "GET", "description": "Check if the server is healthy."},
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
