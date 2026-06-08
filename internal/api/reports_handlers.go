package api

import (
	"errors"
	"net/http"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostReport(w http.ResponseWriter, r *http.Request) {
	var rep models.Report
	if !decodeJSONBody(w, r, &rep) {
		return
	}
	if err := h.reportService.Create(r.Context(), rep); err != nil {
		mapPostReportErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetReports(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	reps, err := h.reportService.List(r.Context(), limit, offset, q, status)
	if err != nil {
		mapListReportErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reps)
}

func (h *Handler) handleAdminUpdateReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}

	if err := h.reportService.Update(r.Context(), body.Status, idStr); err != nil {
		mapUpdateReportErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.reportService.Delete(r.Context(), id); err != nil {
		mapDeleteReportErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostReportErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseNameAndLinkUrlRequired):
		writeJSONError(w, http.StatusBadRequest, "Course name and link URL are required")
	default:
		h.LoggerWithID(r).Error("create report failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteReportErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrReportInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid report id")
	case errors.Is(err, errs.ErrReportNotFound):
		writeJSONError(w, http.StatusNotFound, "Report not found")
	default:
		h.LoggerWithID(r).Error("delete report failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapListReportErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, http.StatusBadRequest, "Limit should be between 1-100 and Offset >= 0")
	case errors.Is(err, errs.ErrReportInvalidStatus):
		writeJSONError(w, http.StatusBadRequest, "Status must be open or resolved")
	default:
		h.LoggerWithID(r).Error("list reports failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateReportErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrReportNotFound):
		writeJSONError(w, http.StatusNotFound, "Report not found")
	case errors.Is(err, errs.ErrReportInvalidID):
		writeJSONError(w, http.StatusBadRequest, "Invalid report id")
	case errors.Is(err, errs.ErrStatusRequired):
		writeJSONError(w, http.StatusBadRequest, "Status is required")
	case errors.Is(err, errs.ErrReportInvalidStatus):
		writeJSONError(w, http.StatusBadRequest, "Status must be open or resolved")
	default:
		h.LoggerWithID(r).Error("update report failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}
