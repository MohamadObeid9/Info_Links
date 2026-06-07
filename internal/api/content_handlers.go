package api

import (
	"net/http"
)

// HandleGetContent fetches all navigation data using a single optimized query.
func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	result, err := h.contentService.Get(r.Context())
	if err != nil {
		h.logger.Error("get content failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
