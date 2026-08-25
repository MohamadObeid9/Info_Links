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

type browseEventBody struct {
	Step string `json:"step"`
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

func (h *Handler) handlePostBrowseEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var body browseEventBody
	if !decodeJSONBody(w, r, &body) {
		return
	}
	if err := h.analyticsService.TrackBrowse(r.Context(), userID, body.Step); err != nil {
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
		writeJSONError(w, r, http.StatusBadRequest, "Range must be 7, 30 or 90")
	case errors.Is(err, errs.ErrAnalyticsInvalidVisitorsSort):
		writeJSONError(w, r, http.StatusBadRequest, "visitors_sort must be clicks or name")
	default:
		h.LoggerWithID(r).Error("get analytics summary failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapAnalyticsTrackErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrAnalyticsInvalidSearchQuery):
		writeJSONError(w, r, http.StatusBadRequest, "Search query is required")
	case errors.Is(err, errs.ErrAnalyticsInvalidBrowseStep):
		writeJSONError(w, r, http.StatusBadRequest, "Browse step must be year or list")
	default:
		h.LoggerWithID(r).Error("track analytics event failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
