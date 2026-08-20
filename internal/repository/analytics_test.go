package repository

import (
	"context"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestAnalyticsRepo(t *testing.T) (AnalyticsRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresAnalyticsRepository(db), mock
}

// analyticsQueryOrder is the order GetSummary runs its aggregations in.
var analyticsQueryOrder = []string{
	analyticsCountsQuery,
	analyticsDailyUniqueVisitsQuery,
	analyticsDailyRosterQuery,
	analyticsTopLinksQuery,
	analyticsTopUsersQuery,
	analyticsTopLinksTodayQuery,
	analyticsVisitorsTodayByClicksQuery,
}

func analyticsQueryNeedsDays(query string) bool {
	switch query {
	case analyticsDailyUniqueVisitsQuery, analyticsDailyRosterQuery, analyticsTopLinksQuery, analyticsTopUsersQuery:
		return true
	default:
		return false
	}
}

func analyticsQueryNeedsVisitorsPaging(query string) bool {
	return query == analyticsVisitorsTodayByClicksQuery || query == analyticsVisitorsTodayByNameQuery
}

func TestAnalyticsRepository_GetSummary(t *testing.T) {
	const days = 30
	linkID := 1
	params := AnalyticsSummaryParams{
		Days:           days,
		VisitorsLimit:  12,
		VisitorsOffset: 0,
		VisitorsSort:   "clicks",
	}

	tests := []struct {
		name    string
		params  AnalyticsSummaryParams
		failAt  int // 1-based index in analyticsQueryOrder, 0 means no failure
		want    models.AnalyticsSummary
		wantErr error
	}{
		{
			name:   "aggregates every metric",
			params: params,
			want: models.AnalyticsSummary{
				TotalStudents:     4,
				StudentsGained7d:  1,
				StudentsGained30d: 2,
				StudentsGained90d: 3,
				ActiveToday:       1,
				ClicksToday:       10,
				DevicesToday:      models.DeviceSplit{Phone: 5, Laptop: 3},
				DailyUniqueVisits: []models.DailyUniqueDay{{Day: "2026-08-18", Users: 12}},
				DailyRoster:       []models.DailyRosterDay{{Day: "2026-08-18", Total: 4}},
				TopLinks:          []models.LinkClickCount{{LinkID: &linkID, Clicks: 9}},
				TopUsers:          []models.UserClickCount{{UserID: 1, Handle: "mohamad_hassan_55", Clicks: 9}},
				TopLinksToday:     []models.LinkClickCount{{LinkID: &linkID, Clicks: 3}},
				VisitorsToday: models.VisitorsTodayPage{
					Visitors: []models.UserClickCount{
						{UserID: 2, Handle: "guest_2", Clicks: 0},
						{UserID: 100, Handle: "extra_visitor_1", Clicks: 0},
						{UserID: 101, Handle: "extra_visitor_2", Clicks: 0},
						{UserID: 102, Handle: "extra_visitor_3", Clicks: 0},
						{UserID: 103, Handle: "extra_visitor_4", Clicks: 0},
						{UserID: 104, Handle: "extra_visitor_5", Clicks: 0},
						{UserID: 105, Handle: "extra_visitor_6", Clicks: 0},
						{UserID: 106, Handle: "extra_visitor_7", Clicks: 0},
						{UserID: 107, Handle: "extra_visitor_8", Clicks: 0},
						{UserID: 108, Handle: "extra_visitor_9", Clicks: 0},
						{UserID: 109, Handle: "extra_visitor_10", Clicks: 0},
						{UserID: 110, Handle: "extra_visitor_11", Clicks: 0},
					},
					HasMore: true,
				},
			},
		},
		{
			name:    "counts query error",
			params:  params,
			failAt:  1,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "daily unique visits query error",
			params:  params,
			failAt:  2,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "daily roster query error",
			params:  params,
			failAt:  3,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "top links query error",
			params:  params,
			failAt:  4,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "top users query error",
			params:  params,
			failAt:  5,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "top links today query error",
			params:  params,
			failAt:  6,
			wantErr: errs.ErrDatabaseDown,
		},
		{
			name:    "visitors today query error",
			params:  params,
			failAt:  7,
			wantErr: errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestAnalyticsRepo(t)
			p := tt.params
			if p.Days == 0 {
				p = params
			}

			for i, query := range analyticsQueryOrder {
				step := i + 1
				if tt.failAt != 0 && step > tt.failAt {
					break
				}

				exp := mock.ExpectQuery(query)
				if analyticsQueryNeedsDays(query) {
					exp = exp.WithArgs(days)
				}
				if analyticsQueryNeedsVisitorsPaging(query) {
					exp = exp.WithArgs(p.VisitorsLimit+1, p.VisitorsOffset)
				}
				if step == tt.failAt {
					exp.WillReturnError(errs.ErrDatabaseDown)
					break
				}
				exp.WillReturnRows(analyticsRowsFor(query, p))
			}

			got, err := repo.GetSummary(context.Background(), p)
			if tt.wantErr != nil {
				assertRepoErr(t, mock, err, tt.wantErr)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAnalyticsRepository_GetSummary_visitorsSortName(t *testing.T) {
	repo, mock := newTestAnalyticsRepo(t)
	params := AnalyticsSummaryParams{
		Days:           7,
		VisitorsLimit:  12,
		VisitorsOffset: 12,
		VisitorsSort:   "name",
	}

	mock.ExpectQuery(analyticsCountsQuery).WillReturnRows(
		sqlmock.NewRows([]string{
			"total_students", "students_gained_7d", "students_gained_30d", "students_gained_90d",
			"active_today", "clicks_today", "phone", "laptop",
		}).AddRow(1, 0, 0, 0, 1, 0, 0, 0),
	)
	mock.ExpectQuery(analyticsDailyUniqueVisitsQuery).WithArgs(7).WillReturnRows(
		sqlmock.NewRows([]string{"day", "users"}),
	)
	mock.ExpectQuery(analyticsDailyRosterQuery).WithArgs(7).WillReturnRows(
		sqlmock.NewRows([]string{"day", "total"}),
	)
	mock.ExpectQuery(analyticsTopLinksQuery).WithArgs(7).WillReturnRows(
		sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}),
	)
	mock.ExpectQuery(analyticsTopUsersQuery).WithArgs(7).WillReturnRows(
		sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}),
	)
	mock.ExpectQuery(analyticsTopLinksTodayQuery).WillReturnRows(
		sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}),
	)
	mock.ExpectQuery(analyticsVisitorsTodayByNameQuery).WithArgs(13, 12).WillReturnRows(
		sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "ali", "ahmad", 1, 0),
	)

	got, err := repo.GetSummary(context.Background(), params)
	assertRepoErr(t, mock, err, nil)
	if len(got.VisitorsToday.Visitors) != 1 || got.VisitorsToday.Visitors[0].Handle != "ali_ahmad_1" {
		t.Fatalf("visitors = %+v", got.VisitorsToday)
	}
}

func analyticsRowsFor(query string, params AnalyticsSummaryParams) *sqlmock.Rows {
	switch query {
	case analyticsCountsQuery:
		return sqlmock.NewRows([]string{
			"total_students", "students_gained_7d", "students_gained_30d", "students_gained_90d",
			"active_today", "clicks_today", "phone", "laptop",
		}).AddRow(4, 1, 2, 3, 1, 10, 5, 3)
	case analyticsDailyUniqueVisitsQuery:
		return sqlmock.NewRows([]string{"day", "users"}).AddRow("2026-08-18", 12)
	case analyticsDailyRosterQuery:
		return sqlmock.NewRows([]string{"day", "total"}).AddRow("2026-08-18", 4)
	case analyticsTopLinksQuery:
		return sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}).AddRow(1, nil, 9)
	case analyticsTopUsersQuery:
		return sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "mohamad", "hassan", 55, 9)
	case analyticsTopLinksTodayQuery:
		return sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}).AddRow(1, nil, 3)
	default:
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(2, "", "", 0, 0)
		for i := 0; i < params.VisitorsLimit; i++ {
			rows.AddRow(100+i, "extra", "visitor", i+1, 0)
		}
		return rows
	}
}
