package api

import (
	"net/http"

	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostLinkClick(w http.ResponseWriter, r *http.Request) {
	var lc models.LinkClick
	if !decodeJSONBody(w, r, &lc) {
		return
	}
	if err := h.linkClickService.Create(r.Context(), lc); err != nil {
		h.LoggerWithID(r).Error("post link click failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetLinkClicks(w http.ResponseWriter, r *http.Request) {
	views, err := h.linkClickService.List(r.Context())
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, views)
}
