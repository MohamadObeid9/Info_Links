package api

import (
	"net/http"

	"infolinks-backend/internal/database"
)

// HandleGetContent fetches all navigation data using a single optimized query.
func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	query := `
		WITH content AS (
			SELECT
				(SELECT COALESCE(json_agg(p ORDER BY display_order ASC), '[]') FROM programs p) as programs,
				(SELECT COALESCE(json_agg(y ORDER BY display_order ASC), '[]') FROM years y) as years,
				(SELECT COALESCE(json_agg(s ORDER BY display_order ASC), '[]') FROM semesters s) as semesters,
				(SELECT COALESCE(json_agg(c ORDER BY display_order ASC), '[]') FROM courses c) as courses,
				(SELECT COALESCE(json_agg(l ORDER BY display_order ASC), '[]') FROM links l WHERE course_id IS NOT NULL) as links,
				(SELECT COALESCE(json_agg(ex ORDER BY display_order ASC), '[]') FROM extra_sections ex) as extra_sections,
				(SELECT COALESCE(json_agg(el ORDER BY display_order ASC), '[]') FROM extra_links el) as extra_links
		)
		SELECT json_build_object(
			'programs', programs,
			'years', years,
			'semesters', semesters,
			'courses', courses,
			'links', links,
			'extra_sections', extra_sections,
			'extra_links', extra_links
		) FROM content;
	`

	var result string
	err := database.DB.QueryRow(query).Scan(&result)
	if err != nil {
		h.logger.Error("database error in HandleGetContent", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}
