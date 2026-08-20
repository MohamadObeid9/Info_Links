package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/service"
)

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
