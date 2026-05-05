package api

import (
	"fmt"
	"net/http"
	"strings"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostContribution(w http.ResponseWriter, r *http.Request) {
	var c models.Contribution
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO contributions (course_name, link_url, note) VALUES ($1, $2, $3)",
		c.CourseName, c.LinkURL, c.Note)
	if err != nil {
		h.logger.Error("db error", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetContributions(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, course_name, link_url, note, status, created_at FROM contributions"
	var args []any
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(course_name ILIKE $%d OR link_url ILIKE $%d OR note ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var contribs []models.Contribution
	for rows.Next() {
		var c models.Contribution
		if err := rows.Scan(&c.ID, &c.CourseName, &c.LinkURL, &c.Note, &c.Status, &c.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		contribs = append(contribs, c)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, contribs)
}

func (h *Handler) handleAdminUpdateContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}
	_, err := database.DB.Exec("UPDATE contributions SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM contributions WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
