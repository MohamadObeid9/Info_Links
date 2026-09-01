package api

import (
	"net/http"
)

const (
	contentCachePublic = "public, max-age=3600, stale-while-revalidate=600"
	contentCacheAdmin  = "private, no-store"
)

// HandleGetContent fetches all navigation data using a single optimized query.
func (h *Handler) handleGetContent(w http.ResponseWriter, r *http.Request) {
	result, err := h.contentService.Get(r.Context())
	if err != nil {
		h.LoggerWithID(r).Error("get content failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", contentCachePublic)
	_, _ = w.Write(result)
}

// handleGetAdminContent is the same payload as GET /api/content, but never
// CDN-cached or served from the in-memory student cache. Admin edits must
// not wait on stale-while-revalidate at the edge or origin RAM.
func (h *Handler) handleGetAdminContent(w http.ResponseWriter, r *http.Request) {
	if err := h.serviceService.FreezeExpired(r.Context()); err != nil {
		h.LoggerWithID(r).Error("freeze expired services failed", "error", err)
	}
	result, err := h.contentService.GetUncached(r.Context())
	if err != nil {
		h.LoggerWithID(r).Error("get content failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", contentCacheAdmin)
	_, _ = w.Write(result)
}

func (h *Handler) invalidateContent() {
	h.contentService.Invalidate()
}
