package api

import (
	"net/http"

	"infolinks-backend/internal/middleware"
)

const (
	contentCachePublic = "public, max-age=3600, stale-while-revalidate=600"
	contentCacheAdmin  = "private, no-store"
)

// handleGetContent fetches all navigation data using a single optimized query.
// A valid admin token gets a no-store response so the browser does not keep it.
func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	result, err := h.contentService.Get(r.Context())
	if err != nil {
		h.LoggerWithID(r).Error("get content failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	cacheControl := contentCachePublic
	if middleware.IsAuthenticatedAdmin(string(h.jwtSecret), r.Header.Get("Authorization")) {
		cacheControl = contentCacheAdmin
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", cacheControl)
	_, _ = w.Write(result)
}

func (h *Handler) invalidateContent() {
	h.contentService.Invalidate()
}
