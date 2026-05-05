package api

import (
	"net/http"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminPostCourse(w http.ResponseWriter, r *http.Request) {
	var c models.Course
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO courses (semester_id, name, code, is_optional, display_order) VALUES ($1, $2, $3, $4, $5)",
		c.SemesterID, c.Name, c.Code, c.IsOptional, c.DisplayOrder)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var c models.Course
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("UPDATE courses SET name = $1, code = $2, semester_id = $3 WHERE id = $4",
		c.Name, c.Code, c.SemesterID, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM courses WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
