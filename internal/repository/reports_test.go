package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func driverValues(args []any) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a
	}
	return vals
}

func newTestReportRepo(t *testing.T) (*PostgresReportRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresReportRepository(db), mock
}

func TestPostgresReportRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		report  models.Report
		execErr error
		err     error
	}{
		{
			name:   "insert report",
			report: models.Report{CourseName: "My Course", LinkURL: "https://test.com", Description: "note"},
		},
		{
			name:    "insert exec error",
			report:  models.Report{CourseName: "My Course", LinkURL: "https://test.com", Description: "note"},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestReportRepo(t)
			exp := mock.ExpectExec(insertReportQuery).
				WithArgs(tt.report.CourseName, tt.report.LinkURL, tt.report.Description)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.report)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Create succeeded, want error %v", tt.err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}

func TestPostgresReportRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "report not found",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrReportNotFound,
		},
		{
			name:    "delete exec error",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept a valid id",
			id:           10,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestReportRepo(t)
			exp := mock.ExpectExec(deleteReportQuery).WithArgs(tt.id)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Delete(context.Background(), tt.id)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Delete succeeded, want error %v", tt.err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}

func TestPostgresReportRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "report not found",
			status:       "open",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrReportNotFound,
		},
		{
			name:    "update exec error",
			status:  "resolved",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid resolved status",
			status:       "resolved",
			id:           10,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestReportRepo(t)
			exp := mock.ExpectExec(updateReportQuery).WithArgs(tt.status, tt.id)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Update(context.Background(), tt.status, tt.id)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Update succeeded, want error %v", tt.err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}

func TestPostgresReportRepository_List(t *testing.T) {
	sampleRow := models.Report{
		ID:          1,
		CourseName:  "Linux",
		LinkURL:     "https://example.com",
		Description: "broken link",
		Status:      "open",
		CreatedAt:   "2024-01-01T00:00:00Z",
	}

	tests := []struct {
		name       string
		limit      int
		offset     int
		q          string
		status     string
		query      string
		queryArgs  []any
		queryErr   error
		rows       [][]any
		rowsErr    error
		err        error
		wantResult []models.Report
	}{
		{
			name:       "list without filters",
			limit:      25,
			offset:     0,
			query:      listReportsNoFilterQuery,
			queryArgs:  []any{25, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Description, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Report{sampleRow},
		},
		{
			name:       "list with search query",
			limit:      10,
			offset:     5,
			q:          "Linux",
			status:     "open",
			query:      listReportsWithQStatusQuery,
			queryArgs:  []any{"%Linux%", "open", 10, 5},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Description, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Report{sampleRow},
		},
		{
			name:       "list with status only",
			limit:      10,
			offset:     10,
			status:     "resolved",
			query:      listReportsWithStatusQuery,
			queryArgs:  []any{"resolved", 10, 10},
			rows:       [][]any{{2, "Go", "https://go.dev", "", "resolved", "2024-02-01T00:00:00Z"}},
			wantResult: []models.Report{{ID: 2, CourseName: "Go", LinkURL: "https://go.dev", Status: "resolved", CreatedAt: "2024-02-01T00:00:00Z"}},
		},
		{
			name:       "list with q only",
			limit:      10,
			offset:     0,
			q:          "Linux",
			query:      listReportsWithQQuery,
			queryArgs:  []any{"%Linux%", 10, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Description, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Report{sampleRow},
		},
		{
			name:      "list query error",
			limit:     10,
			offset:    0,
			status:    "open",
			query:     listReportsWithStatusQuery,
			queryArgs: []any{"open", 10, 0},
			queryErr:  errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:      "list rows error",
			limit:     10,
			offset:    0,
			status:    "open",
			query:     listReportsWithStatusQuery,
			queryArgs: []any{"open", 10, 0},
			rows:      [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Description, sampleRow.Status, sampleRow.CreatedAt}},
			rowsErr:   errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestReportRepo(t)
			exp := mock.ExpectQuery(tt.query).WithArgs(driverValues(tt.queryArgs)...)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				cols := []string{"id", "course_name", "link_url", "description", "status", "created_at"}
				rows := sqlmock.NewRows(cols)
				for _, row := range tt.rows {
					rows.AddRow(driverValues(row)...)
				}
				if tt.rowsErr != nil {
					rows.CloseError(tt.rowsErr)
				}
				exp.WillReturnRows(rows)
			}

			got, err := repo.List(context.Background(), tt.limit, tt.offset, tt.q, tt.status)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("List succeeded, want error %v", tt.err)
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Fatalf("List result: got %+v want %+v", got, tt.wantResult)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}
