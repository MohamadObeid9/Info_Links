package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/service"
)

type searchEventBody struct {
	Query string `json:"query"`
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	visitors := parseAnalyticsVisitorsParams(r)
	summary, err := h.analyticsService.GetSummary(r.Context(), r.URL.Query().Get("range"), visitors)
	if err != nil {
		mapAnalyticsSummaryErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) handleAdminGetAnalyticsActors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := h.analyticsService.ListActors(r.Context(), q.Get("kind"), q.Get("id"), q.Get("range"))
	if err != nil {
		mapAnalyticsActorsErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handlePostSearchEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var body searchEventBody
	if !decodeJSONBody(w, r, &body) {
		return
	}
	if err := h.analyticsService.TrackSearch(r.Context(), userID, body.Query); err != nil {
		mapAnalyticsTrackErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Helpers functions

func parseAnalyticsVisitorsParams(r *http.Request) service.AnalyticsVisitorsParams {
	q := r.URL.Query()
	params := service.AnalyticsVisitorsParams{
		Sort: strings.TrimSpace(q.Get("visitors_sort")),
	}
	if l := q.Get("visitors_limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			params.Limit = parsed
		}
	}
	if o := q.Get("visitors_offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			params.Offset = parsed
		}
	}
	return params
}

func mapAnalyticsSummaryErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrAnalyticsInvalidRange):
		writeJSONError(w, r, http.StatusBadRequest, "Range must be 7, 30, 90, or all")
	case errors.Is(err, errs.ErrAnalyticsInvalidVisitorsSort):
		writeJSONError(w, r, http.StatusBadRequest, "visitors_sort must be clicks or name")
	default:
		h.LoggerWithID(r).Error("get analytics summary failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapAnalyticsActorsErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrAnalyticsInvalidActorKind):
		writeJSONError(w, r, http.StatusBadRequest, "kind must be link, extra_link, course, extra_section, service, favorite, search, or device")
	case errors.Is(err, errs.ErrAnalyticsInvalidActorID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid id")
	case errors.Is(err, errs.ErrAnalyticsInvalidRange):
		writeJSONError(w, r, http.StatusBadRequest, "Range must be today, 7, 30, 90, or all")
	default:
		h.LoggerWithID(r).Error("get analytics actors failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapAnalyticsTrackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrAnalyticsInvalidSearchQuery):
		writeJSONError(w, r, http.StatusBadRequest, "Search query is required")
	default:
		h.LoggerWithID(r).Error("track analytics event failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
