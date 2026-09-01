package api

import (
	"errors"
	"net/http"
	"strings"

	"infolinks-backend/internal/device"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/service"
)

// ── Public Handlers ───────────────────────────────────────────────────────────

func (h *Handler) handleGetServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.serviceService.List(r.Context(), 100, 0, "")
	if err != nil {
		h.LoggerWithID(r).Error("get services failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (h *Handler) handlePostServiceClick(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var click models.ServiceClick
	if !decodeJSONBody(w, r, &click) {
		return
	}
	click.UserID = userID
	if click.ServiceID <= 0 {
		writeJSONError(w, r, http.StatusBadRequest, "service_id is required")
		return
	}
	if click.DeviceType == "" {
		click.DeviceType = device.ClassifyUserAgent(r.UserAgent())
	}
	if click.URL == "" {
		click.URL = strings.TrimSpace(click.ClickedURL)
	}
	service.NormalizeServiceClick(&click)
	if err := h.serviceService.TrackClick(r.Context(), click); err != nil {
		mapPostServiceClickErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetServices(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	services, err := h.serviceService.List(r.Context(), limit, offset, q)
	if err != nil {
		mapListServicesErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (h *Handler) handleAdminPostService(w http.ResponseWriter, r *http.Request) {
	var svc models.Service
	if !decodeJSONBody(w, r, &svc) {
		return
	}
	if err := h.serviceService.Create(r.Context(), svc); err != nil {
		mapPostServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var patch models.ServicePatch
	if !decodeJSONBody(w, r, &patch) {
		return
	}
	if err := h.serviceService.Update(r.Context(), patch, idStr); err != nil {
		mapUpdateServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.serviceService.Delete(r.Context(), idStr); err != nil {
		mapDeleteServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminRenewService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var body struct {
		DurationDays int `json:"duration_days"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}
	if err := h.serviceService.Renew(r.Context(), idStr, body.DurationDays); err != nil {
		mapRenewServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminFreezeService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.serviceService.Freeze(r.Context(), idStr); err != nil {
		mapUpdateServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminUnfreezeService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.serviceService.Unfreeze(r.Context(), idStr); err != nil {
		mapUpdateServiceErr(h, w, r, err)
		return
	}
	h.invalidateContent()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostServiceClickErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrServiceInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid service id")
	default:
		h.LoggerWithID(r).Error("post service click failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapPostServiceErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrServiceTitleRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Service title is required")
	case errors.Is(err, errs.ErrServiceInvalidStatus):
		writeJSONError(w, r, http.StatusBadRequest, "Status must be trial, active, or frozen")
	default:
		h.LoggerWithID(r).Error("create service failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateServiceErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrServiceInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid service id")
	case errors.Is(err, errs.ErrServiceNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Service not found")
	case errors.Is(err, errs.ErrServiceTitleRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Service title is required")
	case errors.Is(err, errs.ErrServiceInvalidStatus):
		writeJSONError(w, r, http.StatusBadRequest, "Status must be trial, active, or frozen")
	case errors.Is(err, errs.ErrServicePatchEmpty):
		writeJSONError(w, r, http.StatusBadRequest, "Service update has no fields")
	default:
		h.LoggerWithID(r).Error("update service failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteServiceErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrServiceInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid service id")
	case errors.Is(err, errs.ErrServiceNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Service not found")
	default:
		h.LoggerWithID(r).Error("delete service failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapListServicesErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, r, http.StatusBadRequest, "Limit should be between 1-100 and offset >= 0")
	default:
		h.LoggerWithID(r).Error("list services failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapRenewServiceErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrServiceInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid service id")
	case errors.Is(err, errs.ErrServiceNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Service not found")
	default:
		h.LoggerWithID(r).Error("renew service failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
