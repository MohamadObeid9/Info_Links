package api

import (
	"net/http"

	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostPageView(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticatedAdmin(string(h.jwtSecret), r.Header.Get("Authorization")) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var pv models.PageView
	if !decodeJSONBody(w, r, &pv) {
		return
	}
	if err := h.pageViewService.Create(r.Context(), pv); err != nil {
		h.LoggerWithID(r).Error("post page view failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetPageViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.pageViewService.List(r.Context())
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, views)
}
