package api

import (
	"net/http"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminPostLink(w http.ResponseWriter, r *http.Request) {
	var l models.Link
	if !decodeJSONBody(w, r, &l) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO links (course_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		l.CourseID, l.Type, l.URL, l.Label, l.Note, l.ContentType, l.DisplayOrder)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var l models.Link
	if !decodeJSONBody(w, r, &l) {
		return
	}
	_, err := database.DB.Exec("UPDATE links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6",
		l.Type, l.URL, l.Label, l.Note, l.ContentType, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM links WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
