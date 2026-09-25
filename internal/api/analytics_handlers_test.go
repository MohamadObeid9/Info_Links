package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/service"
)

type fakeAnalyticsService struct {
	summaryCalls    int
	summaryRange    string
	summaryVisitors service.AnalyticsVisitorsParams
	summary         models.AnalyticsSummary
	summaryErr      error

	actorsCalls int
	actorsKind  string
	actorsID    string
	actorsRange string
	actors      models.AnalyticsActorsResult
	actorsErr   error

	searchCalls int
	searchQuery string
	searchErr   error
}

func (f *fakeAnalyticsService) GetSummary(ctx context.Context, rangeStr string, visitors service.AnalyticsVisitorsParams) (models.AnalyticsSummary, error) {
	f.summaryCalls++
	f.summaryRange = rangeStr
	f.summaryVisitors = visitors
	if f.summaryErr != nil {
		return models.AnalyticsSummary{}, f.summaryErr
	}
	return f.summary, nil
}

func (f *fakeAnalyticsService) ListActors(ctx context.Context, kind, idStr, rangeStr string) (models.AnalyticsActorsResult, error) {
	f.actorsCalls++
	f.actorsKind = kind
	f.actorsID = idStr
	f.actorsRange = rangeStr
	if f.actorsErr != nil {
		return models.AnalyticsActorsResult{}, f.actorsErr
	}
	return f.actors, nil
}

func (f *fakeAnalyticsService) TrackSearch(ctx context.Context, userID int, query string) error {
	f.searchCalls++
	f.searchQuery = query
	return f.searchErr
}

func TestHandleAdminGetAnalyticsSummary(t *testing.T) {
	summary := models.AnalyticsSummary{
		TotalStudents:    120,
		StudentsGained7d: 5,
		ActiveToday:      14,
		ClicksToday:      42,
		DevicesToday:     models.DeviceSplit{Phone: 12, Laptop: 8},
		DailyUniqueVisits: []models.DailyUniqueDay{
			{Day: "2026-08-18", Users: 14},
		},
		TopUsers: []models.UserClickCount{
			{UserID: 7, Handle: "mohamad_hassan_55", Clicks: 9},
		},
		VisitorsToday: models.VisitorsTodayPage{
			Visitors: []models.UserClickCount{{UserID: 7, Handle: "mohamad_hassan_55", Clicks: 2}},
		},
	}

	tests := []struct {
		name         string
		query        string
		summaryErr   error
		statusWanted int
		errMsg       string
		wantRange    string
		wantVisitors service.AnalyticsVisitorsParams
	}{
		{
			name:         "200 with the aggregated summary",
			query:        "?range=30&visitors_limit=12&visitors_offset=0&visitors_sort=clicks",
			statusWanted: http.StatusOK,
			wantRange:    "30",
			wantVisitors: service.AnalyticsVisitorsParams{Limit: 12, Sort: "clicks"},
		},
		{
			name:         "passes an empty range through to the service",
			statusWanted: http.StatusOK,
		},
		{
			name:         "400 on an unsupported range",
			query:        "?range=5",
			summaryErr:   errs.ErrAnalyticsInvalidRange,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Range must be 7, 30 or 90",
			wantRange:    "5",
		},
		{
			name:         "400 on an unsupported visitors sort",
			query:        "?visitors_sort=recent",
			summaryErr:   errs.ErrAnalyticsInvalidVisitorsSort,
			statusWanted: http.StatusBadRequest,
			errMsg:       "visitors_sort must be clicks or name",
			wantVisitors: service.AnalyticsVisitorsParams{Sort: "recent"},
		},
		{
			name:         "500 when the service fails",
			query:        "?range=7",
			summaryErr:   errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantRange:    "7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeAnalytics := &fakeAnalyticsService{summary: summary, summaryErr: tt.summaryErr}
			h := testHandler(t, withAnalytics(fakeAnalytics))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/analytics/summary"+tt.query, nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetAnalyticsSummary(rr, req)

			if fakeAnalytics.summaryCalls != 1 {
				t.Fatalf("summary calls = %d, want 1", fakeAnalytics.summaryCalls)
			}
			if fakeAnalytics.summaryRange != tt.wantRange {
				t.Fatalf("range = %q, want %q", fakeAnalytics.summaryRange, tt.wantRange)
			}
			if fakeAnalytics.summaryVisitors != tt.wantVisitors {
				t.Fatalf("visitors = %+v, want %+v", fakeAnalytics.summaryVisitors, tt.wantVisitors)
			}
			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg != "" {
				var body map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body["error"] != tt.errMsg {
					t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
				}
				return
			}

			var got models.AnalyticsSummary
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got.TotalStudents != summary.TotalStudents || got.ActiveToday != summary.ActiveToday {
				t.Fatalf("summary = %+v, want %+v", got, summary)
			}
			if len(got.TopUsers) != 1 || got.TopUsers[0].Handle != "mohamad_hassan_55" {
				t.Fatalf("top users = %+v, want the seeded student", got.TopUsers)
			}
		})
	}
}

func TestHandlePostSearchEvent(t *testing.T) {
	fake := &fakeAnalyticsService{}
	h := testHandler(t, withAnalytics(fake))
	req := studentRequest(http.MethodPost, "/api/search_events", `{"query":"nfa035"}`)
	rr := httptest.NewRecorder()
	h.handlePostSearchEvent(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if fake.searchCalls != 1 || fake.searchQuery != "nfa035" {
		t.Fatalf("search = %d %q", fake.searchCalls, fake.searchQuery)
	}
}

func TestHandleAdminGetAnalyticsActors(t *testing.T) {
	actors := models.AnalyticsActorsResult{
		Kind:  "link",
		ID:    9,
		Total: 4,
		People: []models.UserClickCount{
			{UserID: 7, Handle: "mohamad_hassan_55", Clicks: 4},
		},
	}

	tests := []struct {
		name         string
		query        string
		actorsErr    error
		statusWanted int
		errMsg       string
		wantKind     string
		wantID       string
		wantRange    string
	}{
		{
			name:         "200 with people and counts",
			query:        "?kind=link&id=9&range=7",
			statusWanted: http.StatusOK,
			wantKind:     "link",
			wantID:       "9",
			wantRange:    "7",
		},
		{
			name:         "400 on bad kind",
			query:        "?kind=widget&id=1&range=7",
			actorsErr:    errs.ErrAnalyticsInvalidActorKind,
			statusWanted: http.StatusBadRequest,
			errMsg:       "kind must be link, extra_link, course, extra_section, service, or favorite",
			wantKind:     "widget",
			wantID:       "1",
			wantRange:    "7",
		},
		{
			name:         "400 on bad id",
			query:        "?kind=link&id=abc&range=7",
			actorsErr:    errs.ErrAnalyticsInvalidActorID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid id",
			wantKind:     "link",
			wantID:       "abc",
			wantRange:    "7",
		},
		{
			name:         "400 on bad range",
			query:        "?kind=link&id=1&range=5",
			actorsErr:    errs.ErrAnalyticsInvalidRange,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Range must be today, 7, 30 or 90",
			wantKind:     "link",
			wantID:       "1",
			wantRange:    "5",
		},
		{
			name:         "500 when the service fails",
			query:        "?kind=service&id=2&range=30",
			actorsErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantKind:     "service",
			wantID:       "2",
			wantRange:    "30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAnalyticsService{actors: actors, actorsErr: tt.actorsErr}
			h := testHandler(t, withAnalytics(fake))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/analytics/actors"+tt.query, nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetAnalyticsActors(rr, req)

			if fake.actorsCalls != 1 {
				t.Fatalf("actors calls = %d, want 1", fake.actorsCalls)
			}
			if fake.actorsKind != tt.wantKind || fake.actorsID != tt.wantID || fake.actorsRange != tt.wantRange {
				t.Fatalf("args = %s/%s/%s, want %s/%s/%s", fake.actorsKind, fake.actorsID, fake.actorsRange, tt.wantKind, tt.wantID, tt.wantRange)
			}
			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg != "" {
				var body map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body["error"] != tt.errMsg {
					t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
				}
				return
			}

			var got models.AnalyticsActorsResult
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got.Total != actors.Total || len(got.People) != 1 || got.People[0].Handle != "mohamad_hassan_55" {
				t.Fatalf("actors = %+v, want %+v", got, actors)
			}
		})
	}
}
