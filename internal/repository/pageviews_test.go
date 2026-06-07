package repository

import (
	"context"
	"errors"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestPageViewRepo(t *testing.T) (PageViewRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresPageViewRepository(db), mock
}

func TestPageViewRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		execErr error
		err     error
		pv      models.PageView
	}{
		{
			name: "insert page view",
			pv:   models.PageView{Page: "home"},
		},
		{
			name:    "insert exec error",
			pv:      models.PageView{Page: "home"},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestPageViewRepo(t)
			exp := mock.ExpectExec(insertPageViewQuery).
				WithArgs(tt.pv.Page)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.pv)
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

func TestPageViewRepository_List(t *testing.T) {
	columns := []string{"id", "page", "visited_at"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []models.PageView
		err      error
	}{
		{
			name: "returns page views",
			rows: sqlmock.NewRows(columns).
				AddRow(1, "home", "2024-01-01T00:00:00Z").
				AddRow(2, "admin", "2024-02-01T00:00:00Z"),
			want: []models.PageView{
				{ID: 1, Page: "home", VisitedAt: "2024-01-01T00:00:00Z"},
				{ID: 2, Page: "admin", VisitedAt: "2024-02-01T00:00:00Z"},
			},
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
			repo, mock := newTestPageViewRepo(t)

			exp := mock.ExpectQuery(GetPageViewQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.List(context.Background())

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("expectations: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}
