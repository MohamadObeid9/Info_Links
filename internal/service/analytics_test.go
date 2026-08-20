package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

type fakeAnalyticsRepo struct {
	calls  int
	params repository.AnalyticsSummaryParams
	result models.AnalyticsSummary
	err    error
}

func (f *fakeAnalyticsRepo) GetSummary(ctx context.Context, params repository.AnalyticsSummaryParams) (models.AnalyticsSummary, error) {
	f.calls++
	f.params = params
	if f.err != nil {
		return models.AnalyticsSummary{}, f.err
	}
	return f.result, nil
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
