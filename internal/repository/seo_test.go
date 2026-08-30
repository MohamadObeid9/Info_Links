package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"infolinks-backend/internal/errs"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestSEORepo(t *testing.T) (SEORepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresSEORepository(db), mock
}

func testSlugFn(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}

func linksQueryForCourseIDs(ids ...int) string {
	args := make([]string, len(ids))
	for i := range ids {
		args[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf(listSEOLinksByCourseIDsQuery, strings.Join(args, ","))
}

func TestSEORepository_GetCoursePageByCode(t *testing.T) {
	placementCols := []string{
		"id", "name", "code", "is_optional",
		"program_id", "program_name", "year_name", "semester_name",
	}
	linkCols := []string{"id", "label", "url", "note", "content_type", "type"}

	tests := []struct {
		name  string
		code  string
		setup func(mock sqlmock.Sqlmock)
		want  *CoursePageData
		err   error
	}{
		{
			name: "empty code",
			code: "  ",
			err:  errs.ErrCourseNotFound,
		},
		{
			name: "course not found",
			code: "nfa999",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getSEOCoursePlacementsQuery).
					WithArgs("nfa999").
					WillReturnRows(sqlmock.NewRows(placementCols))
			},
			err: errs.ErrCourseNotFound,
		},
		{
			name: "placements query error",
			code: "nfa008",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getSEOCoursePlacementsQuery).
					WithArgs("nfa008").
					WillReturnError(errs.ErrDatabaseDown)
			},
			err: errs.ErrDatabaseDown,
		},
		{
			name: "returns course with deduplicated links",
			code: "nfa008",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getSEOCoursePlacementsQuery).
					WithArgs("nfa008").
					WillReturnRows(sqlmock.NewRows(placementCols).
						AddRow(10, "BDD", "nfa008", false, 1, "Génie Info", "L1", "S1"))
				mock.ExpectQuery(linksQueryForCourseIDs(10)).
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(linkCols).
						AddRow(1, "TD 1", "https://a.test", "", "td", "drive").
						AddRow(2, "TD 1", "https://a.test", "", "td", "drive").
						AddRow(3, "Cours", "https://b.test", "note", "cours", "telegram"))
			},
			want: &CoursePageData{
				Code: "nfa008",
				Name: "BDD",
				Placements: []CoursePlacement{{
					CourseID: 10, CourseName: "BDD", Code: "nfa008",
					ProgramID: 1, ProgramName: "Génie Info", YearName: "L1", SemesterName: "S1",
				}},
				Links: []SEOLink{
					{ID: 1, Label: "TD 1", URL: "https://a.test", ContentType: "td", LinkType: "drive"},
					{ID: 3, Label: "Cours", URL: "https://b.test", Note: "note", ContentType: "cours", LinkType: "telegram"},
				},
			},
		},
		{
			name: "links query error",
			code: "nfa008",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(getSEOCoursePlacementsQuery).
					WithArgs("nfa008").
					WillReturnRows(sqlmock.NewRows(placementCols).
						AddRow(10, "BDD", "nfa008", false, 1, "Génie Info", "L1", "S1"))
				mock.ExpectQuery(linksQueryForCourseIDs(10)).
					WithArgs(10).
					WillReturnError(errs.ErrDatabaseDown)
			},
			err: errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestSEORepo(t)
			if tt.setup != nil {
				tt.setup(mock)
			}

			got, err := repo.GetCoursePageByCode(context.Background(), tt.code)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEORepository_ListCourseCodesForSitemap(t *testing.T) {
	columns := []string{"code"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []string
		err      error
	}{
		{
			name: "returns course codes",
			rows: sqlmock.NewRows(columns).AddRow("nfa008").AddRow("nfa010"),
			want: []string{"nfa008", "nfa010"},
		},
		{
			name: "returns empty list",
			rows: sqlmock.NewRows(columns),
			want: nil,
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestSEORepo(t)
			exp := mock.ExpectQuery(listSEOCourseCodesForSitemapQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.ListCourseCodesForSitemap(context.Background())
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEORepository_ListProgramsForSitemap(t *testing.T) {
	columns := []string{"id", "name"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []ProgramSitemapEntry
		err      error
	}{
		{
			name: "returns programs with slugs",
			rows: sqlmock.NewRows(columns).AddRow(1, "Génie Info"),
			want: []ProgramSitemapEntry{{ID: 1, Name: "Génie Info", Slug: testSlugFn("Génie Info")}},
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestSEORepo(t)
			exp := mock.ExpectQuery(listSEOProgramsQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.ListProgramsForSitemap(context.Background(), testSlugFn)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEORepository_ListCoursesIndex(t *testing.T) {
	columns := []string{"code", "name", "program_name"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []CourseIndexEntry
		err      error
	}{
		{
			name: "returns index rows",
			rows: sqlmock.NewRows(columns).AddRow("nfa008", "BDD", "Génie Info"),
			want: []CourseIndexEntry{{Code: "nfa008", Name: "BDD", ProgramName: "Génie Info"}},
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestSEORepo(t)
			exp := mock.ExpectQuery(listSEOCoursesIndexQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.ListCoursesIndex(context.Background())
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEORepository_GetProgramBySlug(t *testing.T) {
	programCols := []string{"id", "name"}
	courseCols := []string{"code", "name"}

	tests := []struct {
		name  string
		slug  string
		setup func(mock sqlmock.Sqlmock)
		want  *ProgramPageData
		err   error
	}{
		{
			name: "empty slug",
			slug: "  ",
			err:  sql.ErrNoRows,
		},
		{
			name: "program not found",
			slug: "missing",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(listSEOProgramsQuery).
					WillReturnRows(sqlmock.NewRows(programCols).AddRow(1, "Génie Info"))
			},
			err: sql.ErrNoRows,
		},
		{
			name: "returns program with courses",
			slug: testSlugFn("Génie Info"),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(listSEOProgramsQuery).
					WillReturnRows(sqlmock.NewRows(programCols).AddRow(1, "Génie Info"))
				mock.ExpectQuery(listSEOProgramCoursesQuery).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(courseCols).AddRow("nfa008", "BDD"))
			},
			want: &ProgramPageData{
				ID:   1,
				Name: "Génie Info",
				Slug: testSlugFn("Génie Info"),
				Courses: []ProgramCourseEntry{
					{Code: "nfa008", Name: "BDD"},
				},
			},
		},
		{
			name: "programs query error",
			slug: "genie-info",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(listSEOProgramsQuery).
					WillReturnError(errs.ErrDatabaseDown)
			},
			err: errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestSEORepo(t)
			if tt.setup != nil {
				tt.setup(mock)
			}

			got, err := repo.GetProgramBySlug(context.Background(), tt.slug, testSlugFn)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
