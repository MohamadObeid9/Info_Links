package api

import (
	"errors"
	"net/http"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostContribution(w http.ResponseWriter, r *http.Request) {
	var contribution models.Contribution
	if !decodeJSONBody(w, r, &contribution) {
		return
	}
	if err := h.contributionService.Create(r.Context(), contribution); err != nil {
		mapPostContributionErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetContributions(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	contributions, err := h.contributionService.List(r.Context(), limit, offset, q, status)
	if err != nil {
		mapListContributionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, contributions)
}

func (h *Handler) handleAdminUpdateContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}

	if err := h.contributionService.Update(r.Context(), body.Status, id); err != nil {
		mapUpdateContributionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.contributionService.Delete(r.Context(), id); err != nil {
		mapDeleteContributionErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostContributionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseNameAndLinkUrlRequired):
		writeJSONError(w, http.StatusBadRequest, "Course name and link URL are required")
	default:
		h.LoggerWithID(r).Error("create contribution failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteContributionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrContributionInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid contribution id")
	case errors.Is(err, errs.ErrContributionNotFound):
		writeJSONError(w, http.StatusNotFound, "Contribution not found")
	default:
		h.LoggerWithID(r).Error("delete contribution failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateContributionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrContributionNotFound):
		writeJSONError(w, http.StatusNotFound, "Contribution not found")
	case errors.Is(err, errs.ErrContributionInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid contribution id")
	case errors.Is(err, errs.ErrStatusRequired):
		writeJSONError(w, http.StatusBadRequest, "Status is required")
	case errors.Is(err, errs.ErrContributionInvalidStatus):
		writeJSONError(w, http.StatusBadRequest, "Status must be pending or approved")
	default:
		h.LoggerWithID(r).Error("update contribution failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapListContributionErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, http.StatusBadRequest, "Limit should be between 1-100 and Offset >= 0")
	case errors.Is(err, errs.ErrContributionInvalidStatus):
		writeJSONError(w, http.StatusBadRequest, "Status must be pending or approved")
	default:
		h.LoggerWithID(r).Error("list contributions failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}
