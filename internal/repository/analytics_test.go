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

var analyticsQueryOrder = []string{
	analyticsCountsQuery,
	analyticsDailyUniqueVisitsQuery,
	analyticsDailyRosterQuery,
	analyticsTopLinksQuery,
	analyticsTopUsersQuery,
	analyticsTopLinksTodayQuery,
	analyticsVisitorsTodayByClicksQuery,
	analyticsTopCoursesQuery,
	analyticsZeroClickCoursesQuery,
	analyticsZeroClickLinksQuery,
	analyticsTopFavoritesQuery,
	analyticsHeatmapQuery,
	analyticsSearchTermsQuery,
}

func analyticsQueryNeedsDays(query string) bool {
	switch query {
	case analyticsCountsQuery,
		analyticsDailyUniqueVisitsQuery,
		analyticsDailyRosterQuery,
		analyticsTopLinksQuery,
		analyticsTopUsersQuery,
		analyticsTopCoursesQuery,
		analyticsZeroClickCoursesQuery,
		analyticsZeroClickLinksQuery,
		analyticsHeatmapQuery,
		analyticsSearchTermsQuery:
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
		failAt  int
		want    models.AnalyticsSummary
		wantErr error
	}{
		{
			name:   "aggregates every metric",
			params: params,
			want: models.AnalyticsSummary{
				TotalStudents:           4,
				StudentsGained7d:        1,
				StudentsGained30d:       2,
				StudentsGained90d:       3,
				ActiveToday:             1,
				ClicksToday:             10,
				DevicesToday:            models.DeviceSplit{Phone: 2, Laptop: 1, Both: 0},
				ActiveInRange:           8,
				ActiveRegisteredInRange: 3,
				ClicksInRange:           40,
				ClickersInRange:         5,
				ClicksPerActive:         5,
				PrevActiveInRange:       6,
				PrevClicksInRange:       30,
				DevicesInRange:          models.DeviceSplit{Phone: 4, Laptop: 3, Both: 1},
				ReturningInRange:        5,
				NewInRange:              3,
				Funnel:                  models.SignupFunnel{Arrivals: 10, SignedUp: 2, StillGuest: 8, GuestsOpen: 20},
				PrevStudentsGained:      1,
				Inbox:                   models.AnalyticsInbox{Reports: 1, Contributions: 2, Feedback: 3},
				Browse:                  models.BrowseDepth{ReachedYear: 7, ReachedList: 4},
				DailyUniqueVisits:       []models.DailyUniqueDay{{Day: "2026-08-18", Users: 12}},
				DailyRoster:             []models.DailyRosterDay{{Day: "2026-08-18", Total: 4}},
				TopLinks:                []models.LinkClickCount{{LinkID: &linkID, Clicks: 9}},
				TopUsers:                []models.UserClickCount{{UserID: 1, Handle: "mohamad_hassan_55", Clicks: 9}},
				TopLinksToday:           []models.LinkClickCount{{LinkID: &linkID, Clicks: 3}},
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
				TopCourses:       []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 12, ProgramName: "Licence Info"}},
				ZeroClickCourses: []models.CourseDemand{{CourseID: 3, Name: "Quiet Course", Code: "QC01", Count: 0, ProgramName: "AISL"}},
				ZeroClickLinks:   []models.DeadLink{{Kind: "link", ID: 4, Label: "Link 1", CourseName: "Quiet Course", ProgramName: "IRSM"}},
				TopFavorites:     []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 6, ProgramName: "Licence Info"}},
				Heatmap:          []models.HeatmapCell{{Dow: 1, Hour: 14, Count: 7}},
				SearchTerms:      []models.SearchTermCount{{Query: "nfa035", Count: 4}},
			},
		},
		{
			name:    "counts query error",
			params:  params,
			failAt:  1,
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
				t.Fatalf("got %+v\nwant %+v", got, tt.want)
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

	mock.ExpectQuery(analyticsCountsQuery).WithArgs(7).WillReturnRows(analyticsRowsFor(analyticsCountsQuery, params))
	mock.ExpectQuery(analyticsDailyUniqueVisitsQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"day", "users"}))
	mock.ExpectQuery(analyticsDailyRosterQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"day", "total"}))
	mock.ExpectQuery(analyticsTopLinksQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}))
	mock.ExpectQuery(analyticsTopUsersQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}))
	mock.ExpectQuery(analyticsTopLinksTodayQuery).WillReturnRows(sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}))
	mock.ExpectQuery(analyticsVisitorsTodayByNameQuery).WithArgs(13, 12).WillReturnRows(
		sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "ali", "ahmad", 1, 0),
	)
	mock.ExpectQuery(analyticsTopCoursesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsZeroClickCoursesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsZeroClickLinksQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"kind", "id", "label", "course_name", "program_name"}))
	mock.ExpectQuery(analyticsTopFavoritesQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsHeatmapQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"dow", "hour", "count"}))
	mock.ExpectQuery(analyticsSearchTermsQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"query", "count"}))

	got, err := repo.GetSummary(context.Background(), params)
	assertRepoErr(t, mock, err, nil)
	if len(got.VisitorsToday.Visitors) != 1 || got.VisitorsToday.Visitors[0].Handle != "ali_ahmad_1" {
		t.Fatalf("visitors = %+v", got.VisitorsToday)
	}
}

func TestAnalyticsRepository_InsertSearchAndBrowse(t *testing.T) {
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectExec(insertSearchEventQuery).WithArgs(7, "nfa035").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insertBrowseEventQuery).WithArgs(7, "year").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.InsertSearch(context.Background(), 7, "nfa035"); err != nil {
		t.Fatalf("InsertSearch: %v", err)
	}
	if err := repo.InsertBrowse(context.Background(), 7, "year"); err != nil {
		t.Fatalf("InsertBrowse: %v", err)
	}
	assertRepoErr(t, mock, nil, nil)
}

func analyticsRowsFor(query string, params AnalyticsSummaryParams) *sqlmock.Rows {
	switch query {
	case analyticsCountsQuery:
		return sqlmock.NewRows([]string{
			"total_students", "students_gained_7d", "students_gained_30d", "students_gained_90d",
			"active_today", "clicks_today", "phone_today", "laptop_today", "both_today",
			"active_in_range", "clicks_in_range", "clickers_in_range",
			"prev_active", "prev_clicks",
			"phone_range", "laptop_range", "both_range",
			"returning", "new_in_range",
			"arrivals", "signed_up", "prev_students_gained", "still_guest", "guests_open",
			"reports", "contributions", "feedback",
			"reached_year", "reached_list",
			"active_registered_in_range",
		}).AddRow(
			4, 1, 2, 3,
			1, 10, 2, 1, 0,
			8, 40, 5,
			6, 30,
			4, 3, 1,
			5, 3,
			10, 2, 1, 8, 20,
			1, 2, 3,
			7, 4,
			3,
		)
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
	case analyticsTopCoursesQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(9, "Réseaux", "NFA035", 12, "Licence Info")
	case analyticsZeroClickCoursesQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(3, "Quiet Course", "QC01", 0, "AISL")
	case analyticsZeroClickLinksQuery:
		return sqlmock.NewRows([]string{"kind", "id", "label", "course_name", "program_name"}).AddRow("link", 4, "Link 1", "Quiet Course", "IRSM")
	case analyticsTopFavoritesQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(9, "Réseaux", "NFA035", 6, "Licence Info")
	case analyticsHeatmapQuery:
		return sqlmock.NewRows([]string{"dow", "hour", "count"}).AddRow(1, 14, 7)
	case analyticsSearchTermsQuery:
		return sqlmock.NewRows([]string{"query", "count"}).AddRow("nfa035", 4)
	default:
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(2, "", "", 0, 0)
		for i := 0; i < params.VisitorsLimit; i++ {
			rows.AddRow(100+i, "extra", "visitor", i+1, 0)
		}
		return rows
	}
}
