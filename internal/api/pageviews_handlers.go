package api

import (
	"net/http"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostPageView(w http.ResponseWriter, r *http.Request) {
	var pv models.PageView
	if !decodeJSONBody(w, r, &pv) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO page_views (page) VALUES ($1)", pv.Page)
	if err != nil {
		h.logger.Error("db error", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetPageViews(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, page, visited_at FROM page_views ORDER BY visited_at DESC")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var views []models.PageView
	for rows.Next() {
		var v models.PageView
		if err := rows.Scan(&v.ID, &v.Page, &v.VisitedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, views)
}
