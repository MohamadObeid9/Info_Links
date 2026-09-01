package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func (h *Handler) handleAdminGetExtraLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.extraLinkService.List(r.Context())
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *Handler) handleAdminPostExtraLink(w http.ResponseWriter, r *http.Request) {
	var link models.ExtraLink
	if !decodeJSONBody(w, r, &link) {
		return
	}
	if err := h.extraLinkService.Create(r.Context(), link); err != nil {
		mapPostExtraLinkErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchExtraLink(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var link models.ExtraLink
	if !decodeJSONBody(w, r, &link) {
		return
	}
	if err := h.extraLinkService.Update(r.Context(), link, idStr); err != nil {
		mapUpdateExtraLinkErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteExtraLink(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.extraLinkService.Delete(r.Context(), idStr); err != nil {
		mapDeleteExtraLinkErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func mapPostExtraLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraLinkURLAndLabelRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Extra link url and label are required")
	case errors.Is(err, errs.ErrExtraLinkInvalidSectionID):
		writeJSONError(w, r, http.StatusBadRequest, "Extra link invalid section id")
	default:
		h.LoggerWithID(r).Error("create extra link failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateExtraLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraLinkInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid extra link id")
	case errors.Is(err, errs.ErrExtraLinkNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Extra link not found")
	default:
		h.LoggerWithID(r).Error("update extra link failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteExtraLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraLinkInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid extra link id")
	case errors.Is(err, errs.ErrExtraLinkNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Extra link not found")
	default:
		h.LoggerWithID(r).Error("delete extra link failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
