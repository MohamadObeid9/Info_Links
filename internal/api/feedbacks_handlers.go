package api

import (
	"errors"
	"net/http"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func (h *Handler) handlePostFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var feedback models.Feedback
	if !decodeJSONBody(w, r, &feedback) {
		return
	}
	feedback.UserID = userID
	if err := h.feedbackService.Create(r.Context(), feedback); err != nil {
		mapPostFeedbackErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetFeedback(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	feedbacks, err := h.feedbackService.List(r.Context(), limit, offset, q, status)
	if err != nil {
		mapListFeedbackErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, feedbacks)
}

func (h *Handler) handleAdminPatchFeedback(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSONBody(w, r, &body) {
		return
	}
	if err := h.feedbackService.Update(r.Context(), body.Status, idStr); err != nil {
		mapUpdateFeedbackErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteFeedback(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.feedbackService.Delete(r.Context(), idStr); err != nil {
		mapDeleteFeedbackErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostFeedbackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrFeedbackCategoryAndRatingRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Category and rating are required")
	case errors.Is(err, errs.ErrFeedbackInvalidRating):
		writeJSONError(w, r, http.StatusBadRequest, "Rating should be between 1 and 5")
	case errors.Is(err, errs.ErrFeedbackInvalidCategory):
		writeJSONError(w, r, http.StatusBadRequest, "Category must be one of the following : ui/ux or content or functionality or performance or accessibility")
	default:
		h.LoggerWithID(r).Error("create feedback failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateFeedbackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrFeedbackNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Feedback not found")
	case errors.Is(err, errs.ErrFeedbackInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid feedback id")
	case errors.Is(err, errs.ErrStatusRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Status is required")
	case errors.Is(err, errs.ErrFeedbackInvalidStatus):
		writeJSONError(w, r, http.StatusBadRequest, "Status must be new or read")
	default:
		h.LoggerWithID(r).Error("update feedback failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteFeedbackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrFeedbackInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid feedback id")
	case errors.Is(err, errs.ErrFeedbackNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Feedback not found")
	default:
		h.LoggerWithID(r).Error("delete feedback failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapListFeedbackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrFeedbackInvalidStatus):
		writeJSONError(w, r, http.StatusBadRequest, "Status must be new or read")
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, r, http.StatusBadRequest, "Limit should be between 1-100 and Offset >= 0")
	default:
		h.LoggerWithID(r).Error("list feedbacks failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
