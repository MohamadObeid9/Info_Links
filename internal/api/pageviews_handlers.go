package api

import (
	"net/http"

	"infolinks-backend/internal/device"
	"infolinks-backend/internal/models"
)

type pageViewBody struct {
	Page   string `json:"page"`
	Device string `json:"device"`
	UserID int    `json:"user_id"`
}

func (h *Handler) handlePostPageView(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var body pageViewBody
	if !decodeJSONBody(w, r, &body) {
		return
	}

	pv := models.PageView{
		Page:       body.Page,
		UserID:     userID,
		DeviceType: device.ClassifyUserAgent(r.UserAgent()),
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
