package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func (h *Handler) handleAdminGetExtraSections(w http.ResponseWriter, r *http.Request) {
	sections, err := h.extraSectionService.List(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, sections)
}

func (h *Handler) handleAdminPostExtraSection(w http.ResponseWriter, r *http.Request) {
	var section models.ExtraSection
	if !decodeJSONBody(w, r, &section) {
		return
	}
	if err := h.extraSectionService.Create(r.Context(), section); err != nil {
		mapPostExtraSectionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchExtraSection(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var section models.ExtraSection
	if !decodeJSONBody(w, r, &section) {
		return
	}
	if err := h.extraSectionService.Update(r.Context(), section, idStr); err != nil {
		mapUpdateExtraSectionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteExtraSection(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.extraSectionService.Delete(r.Context(), idStr); err != nil {
		mapDeleteExtraSectionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func mapPostExtraSectionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraSectionTitleRequired):
		writeJSONError(w, http.StatusBadRequest, "Extra section title is required")
	default:
		h.LoggerWithID(r).Error("create extra section failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateExtraSectionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraSectionInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid extra section id")
	case errors.Is(err, errs.ErrExtraSectionNotFound):
		writeJSONError(w, http.StatusNotFound, "Extra section not found")
	case errors.Is(err, errs.ErrExtraSectionTitleRequired):
		writeJSONError(w, http.StatusBadRequest, "Extra section title is required")
	default:
		h.LoggerWithID(r).Error("update extra section failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteExtraSectionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrExtraSectionInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid extra section id")
	case errors.Is(err, errs.ErrExtraSectionNotFound):
		writeJSONError(w, http.StatusNotFound, "Extra section not found")
	default:
		h.LoggerWithID(r).Error("delete extra section failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}
