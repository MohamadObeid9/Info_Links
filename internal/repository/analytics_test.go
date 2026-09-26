package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

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
	analyticsNewStudentsTodayQuery,
	analyticsTopCoursesQuery,
	analyticsTopServicesQuery,
	analyticsZeroClickCoursesQuery,
	analyticsZeroClickServicesQuery,
	analyticsZeroClickLinksQuery,
	analyticsTopFavoritesQuery,
	analyticsVisitHeatmapQuery,
	analyticsClickHeatmapQuery,
	analyticsSearchTermsQuery,
}

var analyticsAllTimeQueryOrder = []string{
	analyticsCountsAllTimeQuery,
	analyticsWeeklyUniqueVisitsQuery,
	analyticsWeeklyRosterQuery,
	analyticsTopLinksAllTimeQuery,
	analyticsTopUsersAllTimeQuery,
	analyticsTopLinksTodayQuery,
	analyticsVisitorsTodayByClicksQuery,
	analyticsNewStudentsTodayQuery,
	analyticsTopCoursesAllTimeQuery,
	analyticsTopServicesAllTimeQuery,
	analyticsZeroClickCoursesAllTimeQuery,
	analyticsZeroClickServicesAllTimeQuery,
	analyticsZeroClickLinksAllTimeQuery,
	analyticsTopFavoritesQuery,
	analyticsVisitHeatmapAllTimeQuery,
	analyticsClickHeatmapAllTimeQuery,
	analyticsSearchTermsAllTimeQuery,
}

func analyticsSummaryQueryOrder(days int) []string {
	if days == 0 {
		return analyticsAllTimeQueryOrder
	}
	return analyticsQueryOrder
}

func analyticsQueryNeedsDays(query string) bool {
	switch query {
	case analyticsCountsQuery,
		analyticsDailyUniqueVisitsQuery,
		analyticsDailyRosterQuery,
		analyticsTopLinksQuery,
		analyticsTopUsersQuery,
		analyticsTopCoursesQuery,
		analyticsTopServicesQuery,
		analyticsZeroClickCoursesQuery,
		analyticsZeroClickServicesQuery,
		analyticsZeroClickLinksQuery,
		analyticsVisitHeatmapQuery,
		analyticsClickHeatmapQuery,
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
	allTimeParams := AnalyticsSummaryParams{
		Days:           0,
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
				AllTimeVisitors:         20,
				CourseLinks:             12,
				ExtraLinks:              4,
				CoursesWithLinks:        8,
				TotalCourses:            10,
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
				PrevStudentsGained:      1,
				Inbox:                   models.AnalyticsInbox{Reports: 1, Contributions: 2, Feedback: 3},
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
				NewStudentsToday:  []models.UserClickCount{{UserID: 11, Handle: "sara_ali_3", Clicks: 0}},
				TopCourses:        []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 12, ProgramName: "Licence Info"}},
				TopServices:       []models.ServiceDemand{{ServiceID: 3, Title: "Rolita's Soap", Category: "Beauty", Count: 5}},
				ZeroClickCourses:  []models.CourseDemand{{CourseID: 3, Name: "Quiet Course", Code: "QC01", Count: 0, ProgramName: "AISL"}},
				ZeroClickServices: []models.ServiceDemand{{ServiceID: 8, Title: "Testing Service 5", Category: "testing", Count: 0}},
				ZeroClickLinks:    []models.DeadLink{{Kind: "link", ID: 4, Label: "Link 1", CourseName: "Quiet Course", ProgramName: "IRSM"}},
				TopFavorites:      []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 6, ProgramName: "Licence Info"}},
				VisitHeatmap:      []models.HeatmapCell{{Dow: 1, Hour: 14, Count: 7}},
				ClickHeatmap:      []models.HeatmapCell{{Dow: 5, Hour: 21, Count: 1}},
				SearchTerms:       []models.SearchTermCount{{Query: "nfa035", Count: 4}},
			},
		},
		{
			name:   "all-time uses weekly series and zero prev_*",
			params: allTimeParams,
			want: models.AnalyticsSummary{
				TotalStudents:           4,
				AllTimeVisitors:         20,
				CourseLinks:             12,
				ExtraLinks:              4,
				CoursesWithLinks:        8,
				TotalCourses:            10,
				StudentsGained7d:        1,
				StudentsGained30d:       2,
				StudentsGained90d:       3,
				ActiveToday:             1,
				ClicksToday:             10,
				DevicesToday:            models.DeviceSplit{Phone: 2, Laptop: 1, Both: 0},
				ActiveInRange:           20,
				ActiveRegisteredInRange: 18,
				ClicksInRange:           200,
				ClickersInRange:         15,
				ClicksPerActive:         10,
				PrevActiveInRange:       0,
				PrevClicksInRange:       0,
				DevicesInRange:          models.DeviceSplit{Phone: 10, Laptop: 8, Both: 2},
				PrevStudentsGained:      0,
				Inbox:                   models.AnalyticsInbox{Reports: 1, Contributions: 2, Feedback: 3},
				DailyUniqueVisits:       []models.DailyUniqueDay{{Day: "2024-09-30", Users: 5}},
				DailyRoster:             []models.DailyRosterDay{{Day: "2024-09-30", Total: 4}},
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
				NewStudentsToday:  []models.UserClickCount{{UserID: 11, Handle: "sara_ali_3", Clicks: 0}},
				TopCourses:        []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 12, ProgramName: "Licence Info"}},
				TopServices:       []models.ServiceDemand{{ServiceID: 3, Title: "Rolita's Soap", Category: "Beauty", Count: 5}},
				ZeroClickCourses:  []models.CourseDemand{{CourseID: 3, Name: "Quiet Course", Code: "QC01", Count: 0, ProgramName: "AISL"}},
				ZeroClickServices: []models.ServiceDemand{{ServiceID: 8, Title: "Testing Service 5", Category: "testing", Count: 0}},
				ZeroClickLinks:    []models.DeadLink{{Kind: "link", ID: 4, Label: "Link 1", CourseName: "Quiet Course", ProgramName: "IRSM"}},
				TopFavorites:      []models.CourseDemand{{CourseID: 9, Name: "Réseaux", Code: "NFA035", Count: 6, ProgramName: "Licence Info"}},
				VisitHeatmap:      []models.HeatmapCell{{Dow: 1, Hour: 14, Count: 7}},
				ClickHeatmap:      []models.HeatmapCell{{Dow: 5, Hour: 21, Count: 1}},
				SearchTerms:       []models.SearchTermCount{{Query: "nfa035", Count: 4}},
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

			for i, query := range analyticsSummaryQueryOrder(p.Days) {
				step := i + 1
				if tt.failAt != 0 && step > tt.failAt {
					break
				}

				exp := mock.ExpectQuery(query)
				if analyticsQueryNeedsDays(query) {
					exp = exp.WithArgs(p.Days)
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
	mock.ExpectQuery(analyticsNewStudentsTodayQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}))
	mock.ExpectQuery(analyticsTopCoursesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsTopServicesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "category", "count"}))
	mock.ExpectQuery(analyticsZeroClickCoursesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsZeroClickServicesQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "category", "count"}))
	mock.ExpectQuery(analyticsZeroClickLinksQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"kind", "id", "label", "course_name", "program_name"}))
	mock.ExpectQuery(analyticsTopFavoritesQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}))
	mock.ExpectQuery(analyticsVisitHeatmapQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"dow", "hour", "count"}))
	mock.ExpectQuery(analyticsClickHeatmapQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"dow", "hour", "count"}))
	mock.ExpectQuery(analyticsSearchTermsQuery).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"query", "count"}))

	got, err := repo.GetSummary(context.Background(), params)
	assertRepoErr(t, mock, err, nil)
	if len(got.VisitorsToday.Visitors) != 1 || got.VisitorsToday.Visitors[0].Handle != "ali_ahmad_1" {
		t.Fatalf("visitors = %+v", got.VisitorsToday)
	}
}

func TestAnalyticsRepository_InsertSearch(t *testing.T) {
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectExec(insertSearchEventQuery).WithArgs(7, "nfa035").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.InsertSearch(context.Background(), 7, "nfa035"); err != nil {
		t.Fatalf("InsertSearch: %v", err)
	}
	assertRepoErr(t, mock, nil, nil)
}

func TestAnalyticsRepository_ListActors(t *testing.T) {
	since := time.Date(2026, 9, 18, 0, 0, 0, 0, time.Local)
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectQuery(analyticsActorsLinkQuery).
		WithArgs(9, since).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "mohamad", "hassan", 55, 3).
			AddRow(2, "sara", "ali", 3, 2))

	got, err := repo.ListActors(context.Background(), "link", "9", since)
	assertRepoErr(t, mock, err, nil)
	if got.Kind != "link" || got.ID != 9 || got.Total != 5 {
		t.Fatalf("got %+v", got)
	}
	if len(got.People) != 2 || got.People[0].Handle != "mohamad_hassan_55" || got.People[0].Clicks != 3 {
		t.Fatalf("people = %+v", got.People)
	}
}

func TestAnalyticsRepository_ListActors_favorite(t *testing.T) {
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectQuery(analyticsActorsFavoriteQuery).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(8, "ziad", "baroud", 25, 1))

	got, err := repo.ListActors(context.Background(), "favorite", "4", time.Time{})
	assertRepoErr(t, mock, err, nil)
	if got.Total != 1 || len(got.People) != 1 || got.People[0].Handle != "ziad_baroud_25" {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyticsRepository_ListActors_search(t *testing.T) {
	since := time.Date(2026, 9, 18, 0, 0, 0, 0, time.Local)
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectQuery(analyticsActorsSearchQuery).
		WithArgs("delf", since).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(3, "ali", "masmas", 21, 2))

	got, err := repo.ListActors(context.Background(), "search", "delf", since)
	assertRepoErr(t, mock, err, nil)
	if got.Kind != "search" || got.Key != "delf" || got.Total != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyticsRepository_ListActors_device(t *testing.T) {
	since := time.Date(2026, 9, 18, 0, 0, 0, 0, time.Local)
	repo, mock := newTestAnalyticsRepo(t)
	mock.ExpectQuery(analyticsActorsDeviceQuery).
		WithArgs("phone", since).
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "sara", "ali", 3, 1))

	got, err := repo.ListActors(context.Background(), "device", "phone", since)
	assertRepoErr(t, mock, err, nil)
	if got.Kind != "device" || got.Key != "phone" || len(got.People) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyticsRepository_ListActors_invalidKind(t *testing.T) {
	repo, _ := newTestAnalyticsRepo(t)
	_, err := repo.ListActors(context.Background(), "widget", "1", time.Now())
	if !errors.Is(err, errs.ErrAnalyticsInvalidActorKind) {
		t.Fatalf("got %v, want invalid kind", err)
	}
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
			"prev_students_gained",
			"reports", "contributions", "feedback",
			"active_registered_in_range",
			"all_time_visitors",
			"course_links", "extra_links", "courses_with_links", "total_courses",
		}).AddRow(
			4, 1, 2, 3,
			1, 10, 2, 1, 0,
			8, 40, 5,
			6, 30,
			4, 3, 1,
			1,
			1, 2, 3,
			3,
			20,
			12, 4, 8, 10,
		)
	case analyticsCountsAllTimeQuery:
		return sqlmock.NewRows([]string{
			"total_students", "students_gained_7d", "students_gained_30d", "students_gained_90d",
			"active_today", "clicks_today", "phone_today", "laptop_today", "both_today",
			"active_in_range", "clicks_in_range", "clickers_in_range",
			"prev_active", "prev_clicks",
			"phone_range", "laptop_range", "both_range",
			"prev_students_gained",
			"reports", "contributions", "feedback",
			"active_registered_in_range",
			"all_time_visitors",
			"course_links", "extra_links", "courses_with_links", "total_courses",
		}).AddRow(
			4, 1, 2, 3,
			1, 10, 2, 1, 0,
			20, 200, 15,
			0, 0,
			10, 8, 2,
			0,
			1, 2, 3,
			18,
			20,
			12, 4, 8, 10,
		)
	case analyticsDailyUniqueVisitsQuery:
		return sqlmock.NewRows([]string{"day", "users"}).AddRow("2026-08-18", 12)
	case analyticsWeeklyUniqueVisitsQuery:
		return sqlmock.NewRows([]string{"day", "users"}).AddRow("2024-09-30", 5)
	case analyticsDailyRosterQuery:
		return sqlmock.NewRows([]string{"day", "total"}).AddRow("2026-08-18", 4)
	case analyticsWeeklyRosterQuery:
		return sqlmock.NewRows([]string{"day", "total"}).AddRow("2024-09-30", 4)
	case analyticsTopLinksQuery, analyticsTopLinksAllTimeQuery:
		return sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}).AddRow(1, nil, 9)
	case analyticsTopUsersQuery, analyticsTopUsersAllTimeQuery:
		return sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(1, "mohamad", "hassan", 55, 9)
	case analyticsTopLinksTodayQuery:
		return sqlmock.NewRows([]string{"link_id", "extra_link_id", "clicks"}).AddRow(1, nil, 3)
	case analyticsNewStudentsTodayQuery:
		return sqlmock.NewRows([]string{"id", "first_name", "last_name", "number", "clicks"}).
			AddRow(11, "sara", "ali", 3, 0)
	case analyticsTopCoursesQuery, analyticsTopCoursesAllTimeQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(9, "Réseaux", "NFA035", 12, "Licence Info")
	case analyticsTopServicesQuery, analyticsTopServicesAllTimeQuery:
		return sqlmock.NewRows([]string{"id", "title", "category", "count"}).AddRow(3, "Rolita's Soap", "Beauty", 5)
	case analyticsZeroClickCoursesQuery, analyticsZeroClickCoursesAllTimeQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(3, "Quiet Course", "QC01", 0, "AISL")
	case analyticsZeroClickServicesQuery, analyticsZeroClickServicesAllTimeQuery:
		return sqlmock.NewRows([]string{"id", "title", "category", "count"}).AddRow(8, "Testing Service 5", "testing", 0)
	case analyticsZeroClickLinksQuery, analyticsZeroClickLinksAllTimeQuery:
		return sqlmock.NewRows([]string{"kind", "id", "label", "course_name", "program_name"}).AddRow("link", 4, "Link 1", "Quiet Course", "IRSM")
	case analyticsTopFavoritesQuery:
		return sqlmock.NewRows([]string{"id", "name", "code", "count", "program_name"}).AddRow(9, "Réseaux", "NFA035", 6, "Licence Info")
	case analyticsVisitHeatmapQuery, analyticsVisitHeatmapAllTimeQuery:
		return sqlmock.NewRows([]string{"dow", "hour", "count"}).AddRow(1, 14, 7)
	case analyticsClickHeatmapQuery, analyticsClickHeatmapAllTimeQuery:
		return sqlmock.NewRows([]string{"dow", "hour", "count"}).AddRow(5, 21, 1)
	case analyticsSearchTermsQuery, analyticsSearchTermsAllTimeQuery:
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
