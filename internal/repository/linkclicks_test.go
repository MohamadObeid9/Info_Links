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

func newTestLinkClickRepo(t *testing.T) (LinkClickRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresLinkClickRepository(db), mock
}

func TestLinkClickRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		execErr error
		err     error
		lc      models.LinkClick
	}{
		{
			name: "insert normal link click",
			lc:   models.LinkClick{LinkID: &[]int{1}[0], UserID: 7},
		},
		{
			name: "insert extra link click",
			lc:   models.LinkClick{ExtraLinkID: &[]int{1}[0], UserID: 7},
		},
		{
			name:    "insert exec error",
			lc:      models.LinkClick{LinkID: &[]int{1}[0], UserID: 7},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestLinkClickRepo(t)
			exp := mock.ExpectExec(insertLinkClickQuery).
				WithArgs(tt.lc.LinkID, tt.lc.ExtraLinkID, tt.lc.UserID)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.lc)
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

func TestLinkClickRepository_List(t *testing.T) {
	columns := []string{"id", "link_id", "extra_link_id", "clicked_at"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []models.LinkClick
		err      error
	}{
		{
			name: "returns link clicks",
			rows: sqlmock.NewRows(columns).
				AddRow(1, 42, nil, "2024-01-01T00:00:00Z").
				AddRow(2, 99, nil, "2024-02-01T00:00:00Z"),
			want: []models.LinkClick{
				{ID: 1, LinkID: &[]int{42}[0], ExtraLinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
				{ID: 2, LinkID: &[]int{99}[0], ExtraLinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
			},
		},
		{
			name: "returns extra link clicks",
			rows: sqlmock.NewRows(columns).
				AddRow(1, nil, 42, "2024-01-01T00:00:00Z").
				AddRow(2, nil, 99, "2024-02-01T00:00:00Z"),
			want: []models.LinkClick{
				{ID: 1, ExtraLinkID: &[]int{42}[0], LinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
				{ID: 2, ExtraLinkID: &[]int{99}[0], LinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
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
			repo, mock := newTestLinkClickRepo(t)

			exp := mock.ExpectQuery(GetLinkClickQuery)
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
