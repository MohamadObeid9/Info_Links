package api

import (
	"net/http"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostLinkClick(w http.ResponseWriter, r *http.Request) {
	var lc models.LinkClick
	if !decodeJSONBody(w, r, &lc) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO link_clicks (link_id) VALUES ($1)", lc.LinkID)
	if err != nil {
		h.logger.Error("db error", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetLinkClicks(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, link_id, clicked_at FROM link_clicks WHERE link_id IS NOT NULL ORDER BY clicked_at DESC")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var clicks []models.LinkClick
	for rows.Next() {
		var c models.LinkClick
		if err := rows.Scan(&c.ID, &c.LinkID, &c.ClickedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		clicks = append(clicks, c)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, clicks)
}
