package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostReport(w http.ResponseWriter, r *http.Request) {
	var rep models.Report
	if !decodeJSONBody(w, r, &rep) {
		return
	}
	if err := h.reportService.Create(r.Context(), rep); err != nil {
		h.logger.Error("create report failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

/*
Legacy implementation kept for before/after comparison.

func HandlePostReport(w http.ResponseWriter, r *http.Request) {
	var rep models.Report
	if !decodeJSONBody(w, r, &rep) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)",
		rep.CourseName, rep.LinkURL, rep.Description)
	if err != nil {
		logger.Error("db error", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}
*/

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetReports(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, course_name, link_url, description, status, created_at FROM reports"
	var args []any
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(course_name ILIKE $%d OR description ILIKE $%d OR link_url ILIKE $%d)", argIdx, argIdx, argIdx))
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
	var reps []models.Report
	for rows.Next() {
		var rep models.Report
		if err := rows.Scan(&rep.ID, &rep.CourseName, &rep.LinkURL, &rep.Description, &rep.Status, &rep.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		reps = append(reps, rep)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, reps)
}

func (h *Handler) handleAdminUpdateReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid report id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}
	res, err := database.DB.Exec("UPDATE reports SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeJSONError(w, http.StatusNotFound, "Report not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM reports WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
