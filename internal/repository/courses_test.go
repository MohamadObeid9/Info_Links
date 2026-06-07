package repository

import (
	"context"
	"errors"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestCourseRepo(t *testing.T) (CourseRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresCourseRepository(db), mock
}

func assertRepoErr(t *testing.T, mock sqlmock.Sqlmock, err, want error) {
	t.Helper()
	if want != nil {
		if !errors.Is(err, want) {
			t.Fatalf("got %v, want %v", err, want)
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCourseRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		resultErr    error
		rowsAffected int64
		err          error
	}{
		{
			name:         "course not found",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrCourseNotFound,
		},
		{
			name:    "delete exec error",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:      "rows affected error",
			id:        10,
			resultErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "accept a valid id",
			id:           10,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestCourseRepo(t)
			exp := mock.ExpectExec(deleteCourseQuery).WithArgs(tt.id)
			switch {
			case tt.execErr != nil:
				exp.WillReturnError(tt.execErr)
			case tt.resultErr != nil:
				exp.WillReturnResult(sqlmock.NewErrorResult(tt.resultErr))
			default:
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Delete(context.Background(), tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestCourseRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		course  models.Course
		execErr error
		err     error
	}{
		{
			name:   "insert course",
			course: models.Course{SemesterID: 3, Name: "BDD", Code: "nfa008", IsOptional: false, DisplayOrder: 55},
		},
		{
			name:    "insert exec error",
			course:  models.Course{SemesterID: 3, Name: "BDD", Code: "nfa008", IsOptional: false, DisplayOrder: 55},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestCourseRepo(t)
			exp := mock.ExpectExec(insertCourseQuery).
				WithArgs(tt.course.SemesterID, tt.course.Name, tt.course.Code, tt.course.IsOptional, tt.course.DisplayOrder)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.course)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestCourseRepository_GetByID(t *testing.T) {
	columns := []string{"id", "semester_id", "name", "code", "is_optional", "display_order"}

	tests := []struct {
		name     string
		id       int
		queryErr error
		rows     *sqlmock.Rows
		want     models.Course
		err      error
	}{
		{
			name: "course not found",
			id:   1000000,
			rows: sqlmock.NewRows(columns),
			want: models.Course{},
			err:  errs.ErrCourseNotFound,
		},
		{
			name:     "query error",
			id:       10,
			queryErr: errs.ErrDatabaseDown,
			want:     models.Course{},
			err:      errs.ErrDatabaseDown,
		},
		{
			name: "returns course",
			id:   10,
			rows: sqlmock.NewRows(columns).AddRow(10, 3, "Reseaux", "NFA009", false, 55),
			want: models.Course{
				ID:           10,
				SemesterID:   3,
				Name:         "Reseaux",
				Code:         "NFA009",
				IsOptional:   false,
				DisplayOrder: 55,
			},
		},
		{
			name: "returns optional course",
			id:   11,
			rows: sqlmock.NewRows(columns).AddRow(11, 3, "Elective", "OPT101", true, 0),
			want: models.Course{
				ID:           11,
				SemesterID:   3,
				Name:         "Elective",
				Code:         "OPT101",
				IsOptional:   true,
				DisplayOrder: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestCourseRepo(t)

			exp := mock.ExpectQuery(getCourseByIDQuery).WithArgs(tt.id)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.GetByID(context.Background(), tt.id)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCourseRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		course       models.Course
		id           int
		execErr      error
		resultErr    error
		rowsAffected int64
		err          error
	}{
		{
			name:         "course not found",
			id:           10,
			course:       models.Course{Name: "blabla", Code: "xhahafha55", IsOptional: false, SemesterID: 33},
			rowsAffected: 0,
			err:          errs.ErrCourseNotFound,
		},
		{
			name:    "update exec error",
			id:      10,
			course:  models.Course{Name: "Reseaux", Code: "NFA009", IsOptional: false, SemesterID: 3},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:      "rows affected error",
			id:        10,
			course:    models.Course{Name: "Reseaux", Code: "NFA009", IsOptional: false, SemesterID: 3},
			resultErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid course",
			id:           10,
			course:       models.Course{Name: "Reseaux", Code: "NFA009", IsOptional: true, SemesterID: 3},
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestCourseRepo(t)
			exp := mock.ExpectExec(updateCourseQuery).
				WithArgs(tt.course.Name, tt.course.Code, tt.course.SemesterID, tt.course.IsOptional, tt.id)
			switch {
			case tt.execErr != nil:
				exp.WillReturnError(tt.execErr)
			case tt.resultErr != nil:
				exp.WillReturnResult(sqlmock.NewErrorResult(tt.resultErr))
			default:
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Update(context.Background(), tt.course, tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}
