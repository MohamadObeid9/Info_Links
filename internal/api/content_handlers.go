package api

import (
	"net/http"
)

// HandleGetContent fetches all navigation data using a single optimized query.
func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	if err := h.serviceService.FreezeExpired(r.Context()); err != nil {
		h.LoggerWithID(r).Error("freeze expired services failed", "error", err)
	}
	result, err := h.contentService.Get(r.Context())
	if err != nil {
		h.LoggerWithID(r).Error("get content failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=600")
	_, _ = w.Write(result)
}
