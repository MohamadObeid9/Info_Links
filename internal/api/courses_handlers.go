package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminPostCourse(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	if !decodeJSONBody(w, r, &course) {
		return
	}
	if err := h.courseService.Create(r.Context(), course); err != nil {
		mapPostCourseErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminPatchCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var patch models.CoursePatch
	if !decodeJSONBody(w, r, &patch) {
		return
	}

	if err := h.courseService.Update(r.Context(), patch, idStr); err != nil {
		mapUpdateCourseErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAdminDeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if err := h.courseService.Delete(r.Context(), idStr); err != nil {
		mapDeleteCourseErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Helpers functions

func mapPostCourseErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseCodeAndNameRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Course code and course name are required")
	case errors.Is(err, errs.ErrCourseInvalidSemestreID):
		writeJSONError(w, r, http.StatusBadRequest, "Course invalid semestre id ")
	default:
		h.LoggerWithID(r).Error("create course failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapDeleteCourseErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Course invalid id")
	case errors.Is(err, errs.ErrCourseNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Course not found")
	default:
		h.LoggerWithID(r).Error("delete course failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUpdateCourseErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Course invalid id")
	case errors.Is(err, errs.ErrCourseNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Course not found")
	case errors.Is(err, errs.ErrCourseInvalidSemestreID):
		writeJSONError(w, r, http.StatusBadRequest, "Course invalid semestre id ")
	case errors.Is(err, errs.ErrCoursePatchEmpty):
		writeJSONError(w, r, http.StatusBadRequest, "Course invalid update parameters")
	case errors.Is(err, errs.ErrCourseCodeAndNameRequired):
		writeJSONError(w, r, http.StatusBadRequest, "Course code and course name are required")
	default:
		h.LoggerWithID(r).Error("update course failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
