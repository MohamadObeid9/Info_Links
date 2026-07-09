package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostLinkClick(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticatedAdmin(string(h.jwtSecret), r.Header.Get("Authorization")) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var lc models.LinkClick
	if !decodeJSONBody(w, r, &lc) {
		return
	}
	if err := h.linkClickService.Create(r.Context(), lc); err != nil {
		mapPostLinkClickErr(h, w, r, err)
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

// Helpers functions

func mapPostLinkClickErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrLinkClickLinkIDAndExtraLinkIDRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Link id and extra link id are required")
	case errors.Is(err, errs.ErrLinkClickLinkIDAndExtraLinkIDSet):
		writeJSONError(w, r, http.StatusBadRequest, "Link id and extra link id cannot be set at the same time")
	default:
		h.LoggerWithID(r).Error("post link click failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
