package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminPostLink(w http.ResponseWriter, r *http.Request) {
	var link models.Link
	if !decodeJSONBody(w, r, &link) {
		return
	}
	if err := h.linkService.Create(r.Context(), link); err != nil {
		mapPostLinkErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchLink(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var link models.Link
	if !decodeJSONBody(w, r, &link) {
		return
	}
	if err := h.linkService.Update(r.Context(), link, idStr); err != nil {
		mapUpdateLinkErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteLink(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.linkService.Delete(r.Context(), idStr); err != nil {
		mapDeleteLinkErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrLinkURLAndLabelRequired):
		writeJSONError(w, http.StatusBadRequest, "Link url and link label are required")
	default:
		h.LoggerWithID(r).Error("create link failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrLinkInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid link id")
	case errors.Is(err, errs.ErrLinkNotFound):
		writeJSONError(w, http.StatusNotFound, "Link not found")
	default:
		h.LoggerWithID(r).Error("delete link failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrLinkNotFound):
		writeJSONError(w, http.StatusNotFound, "Link not found")
	case errors.Is(err, errs.ErrLinkInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid link id")
	default:
		h.LoggerWithID(r).Error("update link failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}
