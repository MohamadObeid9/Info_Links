package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

type fakeAnalyticsRepo struct {
	calls  int
	params repository.AnalyticsSummaryParams
	result models.AnalyticsSummary
	err    error

	searchCalls int
	searchQuery string
	searchErr   error

	actorsCalls int
	actorsKind  string
	actorsID    string
	actorsSince time.Time
	actors      models.AnalyticsActorsResult
	actorsErr   error
}

func (f *fakeAnalyticsRepo) GetSummary(ctx context.Context, params repository.AnalyticsSummaryParams) (models.AnalyticsSummary, error) {
	f.calls++
	f.params = params
	if f.err != nil {
		return models.AnalyticsSummary{}, f.err
	}
	return f.result, nil
}

func (f *fakeAnalyticsRepo) InsertSearch(ctx context.Context, userID int, query string) error {
	f.searchCalls++
	f.searchQuery = query
	return f.searchErr
}

func (f *fakeAnalyticsRepo) ListActors(ctx context.Context, kind, key string, since time.Time) (models.AnalyticsActorsResult, error) {
	f.actorsCalls++
	f.actorsKind = kind
	f.actorsID = key
	f.actorsSince = since
	if f.actorsErr != nil {
		return models.AnalyticsActorsResult{}, f.actorsErr
	}
	return f.actors, nil
}

func TestAnalyticsService_GetSummary(t *testing.T) {
	summary := models.AnalyticsSummary{TotalStudents: 4, ActiveToday: 1}

	tests := []struct {
		name      string
		rangeStr  string
		visitors  AnalyticsVisitorsParams
		repoErr   error
		wantCalls int
		wantDays  int
		want      models.AnalyticsSummary
		wantErr   error
		wantLimit int
		wantSort  string
	}{
		{
			name:      "defaults to 7 days and visitor paging defaults",
			wantCalls: 1,
			wantDays:  7,
			want:      summary,
			wantLimit: defaultVisitorsLimit,
			wantSort:  "clicks",
		},
		{
			name:      "accepts 30 days",
			rangeStr:  "30",
			wantCalls: 1,
			wantDays:  30,
			want:      summary,
			wantLimit: defaultVisitorsLimit,
			wantSort:  "clicks",
		},
		{
			name:      "accepts 90 days",
			rangeStr:  " 90 ",
			wantCalls: 1,
			wantDays:  90,
			want:      summary,
			wantLimit: defaultVisitorsLimit,
			wantSort:  "clicks",
		},
		{
			name:      "accepts all time (Days=0)",
			rangeStr:  "all",
			wantCalls: 1,
			wantDays:  0,
			want:      summary,
			wantLimit: defaultVisitorsLimit,
			wantSort:  "clicks",
		},
		{
			name: "passes visitor paging params",
			visitors: AnalyticsVisitorsParams{
				Limit:  24,
				Offset: 12,
				Sort:   "name",
			},
			wantCalls: 1,
			wantDays:  7,
			want:      summary,
			wantLimit: 24,
			wantSort:  "name",
		},
		{
			name:     "rejects an unsupported range",
			rangeStr: "45",
			wantErr:  errs.ErrAnalyticsInvalidRange,
		},
		{
			name: "rejects an unsupported visitors sort",
			visitors: AnalyticsVisitorsParams{
				Sort: "recent",
			},
			wantErr: errs.ErrAnalyticsInvalidVisitorsSort,
		},
		{
			name:      "wraps a repo error",
			rangeStr:  "7",
			repoErr:   errs.ErrDatabaseDown,
			wantCalls: 1,
			wantErr:   errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAnalyticsRepo{result: summary, err: tt.repoErr}
			svc := NewAnalyticsService(repo)

			got, err := svc.GetSummary(context.Background(), tt.rangeStr, tt.visitors)

			if repo.calls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.calls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetSummary: %v", err)
			}
			if repo.params.Days != tt.wantDays {
				t.Fatalf("repo days = %d, want %d", repo.params.Days, tt.wantDays)
			}
			if repo.params.VisitorsLimit != tt.wantLimit {
				t.Fatalf("repo visitors limit = %d, want %d", repo.params.VisitorsLimit, tt.wantLimit)
			}
			if repo.params.VisitorsSort != tt.wantSort {
				t.Fatalf("repo visitors sort = %q, want %q", repo.params.VisitorsSort, tt.wantSort)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAnalyticsService_TrackSearch(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		repoErr   error
		wantCalls int
		wantQuery string
		wantErr   error
	}{
		{name: "normalizes and stores", query: "  NFA035  ", wantCalls: 1, wantQuery: "nfa035"},
		{name: "rejects empty", query: "   ", wantErr: errs.ErrAnalyticsInvalidSearchQuery},
		{name: "wraps repo error", query: "algo", repoErr: errs.ErrDatabaseDown, wantCalls: 1, wantQuery: "algo", wantErr: errs.ErrDatabaseDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAnalyticsRepo{searchErr: tt.repoErr}
			svc := NewAnalyticsService(repo)
			err := svc.TrackSearch(context.Background(), 7, tt.query)
			if repo.searchCalls != tt.wantCalls {
				t.Fatalf("calls = %d, want %d", repo.searchCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("TrackSearch: %v", err)
			}
			if repo.searchQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", repo.searchQuery, tt.wantQuery)
			}
		})
	}
}

func TestAnalyticsService_ListActors(t *testing.T) {
	result := models.AnalyticsActorsResult{
		Kind:  "link",
		ID:    9,
		Total: 5,
		People: []models.UserClickCount{
			{UserID: 1, Handle: "mohamad_hassan_55", Clicks: 3},
			{UserID: 2, Handle: "sara_ali_3", Clicks: 2},
		},
	}

	tests := []struct {
		name      string
		kind      string
		id        string
		rangeStr  string
		repoErr   error
		wantCalls int
		wantKind  string
		wantID    string
		wantZero  bool // favorites ignore since
		wantErr   error
	}{
		{name: "link in range", kind: "link", id: "9", rangeStr: "7", wantCalls: 1, wantKind: "link", wantID: "9"},
		{name: "today start of day", kind: "course", id: "3", rangeStr: "today", wantCalls: 1, wantKind: "course", wantID: "3"},
		{name: "favorite ignores range", kind: "favorite", id: "4", rangeStr: "90", wantCalls: 1, wantKind: "favorite", wantID: "4", wantZero: true},
		{name: "all time zero since", kind: "link", id: "9", rangeStr: "all", wantCalls: 1, wantKind: "link", wantID: "9", wantZero: true},
		{name: "search by query", kind: "search", id: "NFA035", rangeStr: "7", wantCalls: 1, wantKind: "search", wantID: "nfa035"},
		{name: "device phone", kind: "device", id: "phone", rangeStr: "30", wantCalls: 1, wantKind: "device", wantID: "phone"},
		{name: "rejects bad id", kind: "link", id: "0", wantErr: errs.ErrAnalyticsInvalidActorID},
		{name: "rejects bad search", kind: "search", id: "  ", rangeStr: "7", wantErr: errs.ErrAnalyticsInvalidActorID},
		{name: "rejects bad device", kind: "device", id: "tablet", rangeStr: "7", wantErr: errs.ErrAnalyticsInvalidActorID},
		{name: "rejects bad kind", kind: "widget", id: "1", rangeStr: "7", repoErr: errs.ErrAnalyticsInvalidActorKind, wantCalls: 1, wantKind: "widget", wantID: "1", wantErr: errs.ErrAnalyticsInvalidActorKind},
		{name: "rejects bad range", kind: "link", id: "1", rangeStr: "5", wantErr: errs.ErrAnalyticsInvalidRange},
		{name: "wraps repo error", kind: "service", id: "2", rangeStr: "30", repoErr: errs.ErrDatabaseDown, wantCalls: 1, wantKind: "service", wantID: "2", wantErr: errs.ErrDatabaseDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAnalyticsRepo{actors: result, actorsErr: tt.repoErr}
			svc := NewAnalyticsService(repo)
			got, err := svc.ListActors(context.Background(), tt.kind, tt.id, tt.rangeStr)
			if repo.actorsCalls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.actorsCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ListActors: %v", err)
			}
			if repo.actorsKind != tt.wantKind || repo.actorsID != tt.wantID {
				t.Fatalf("repo kind/id = %s/%s, want %s/%s", repo.actorsKind, repo.actorsID, tt.wantKind, tt.wantID)
			}
			if tt.wantZero {
				if !repo.actorsSince.IsZero() {
					t.Fatalf("favorite since = %v, want zero", repo.actorsSince)
				}
			} else if repo.actorsSince.IsZero() {
				t.Fatal("expected non-zero since")
			}
			if !reflect.DeepEqual(got, result) {
				t.Fatalf("got %+v, want %+v", got, result)
			}
		})
	}
}
