package api

import (
	"fmt"
	"net/http"
	"strings"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostFeedback(w http.ResponseWriter, r *http.Request) {
	var f models.Feedback
	if !decodeJSONBody(w, r, &f) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO feedback (category, rating, message) VALUES ($1, $2, $3)",
		f.Category, f.Rating, f.Message)
	if err != nil {
		h.logger.Error("db error", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetFeedback(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, category, rating, message, status, created_at FROM feedback"
	var args []any
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(category ILIKE $%d OR message ILIKE $%d)", argIdx, argIdx))
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
	var feed []models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.ID, &f.Category, &f.Rating, &f.Message, &f.Status, &f.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		feed = append(feed, f)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, feed)
}

func (h *Handler) handleAdminPatchFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}
	_, err := database.DB.Exec("UPDATE feedback SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM feedback WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
