package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestExtraSectionRepo(t *testing.T) (ExtraSectionRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresExtraSectionRepository(db), mock
}

func TestExtraSectionRepository_List(t *testing.T) {
	columns := []string{"id", "title", "icon", "display_order"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []models.ExtraSection
		err      error
	}{
		{
			name: "returns extra sections",
			rows: sqlmock.NewRows(columns).AddRow(1, "Python Resources", "🐍", 0),
			want: []models.ExtraSection{{ID: 1, Title: "Python Resources", Icon: "🐍", DisplayOrder: 0}},
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
			repo, mock := newTestExtraSectionRepo(t)
			exp := mock.ExpectQuery(listExtraSectionsQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.List(context.Background())
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

func TestExtraSectionRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		section models.ExtraSection
		execErr error
		err     error
	}{
		{
			name:    "insert extra section",
			section: models.ExtraSection{Title: "Python Resources", Icon: "🐍", DisplayOrder: 0},
		},
		{
			name:    "insert exec error",
			section: models.ExtraSection{Title: "Python Resources", Icon: "🐍", DisplayOrder: 0},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestExtraSectionRepo(t)
			exp := mock.ExpectExec(insertExtraSectionQuery).
				WithArgs(tt.section.Title, tt.section.Icon, tt.section.DisplayOrder)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.section)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestExtraSectionRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		section      models.ExtraSection
		id           int
		execErr      error
		resultErr    error
		rowsAffected int64
		err          error
	}{
		{
			name:         "extra section not found",
			id:           10,
			section:      models.ExtraSection{Title: "Updated", Icon: "📁"},
			rowsAffected: 0,
			err:          errs.ErrExtraSectionNotFound,
		},
		{
			name:    "update exec error",
			id:      10,
			section: models.ExtraSection{Title: "Updated", Icon: "📁"},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid section",
			id:           10,
			section:      models.ExtraSection{Title: "Updated", Icon: "📁"},
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestExtraSectionRepo(t)
			exp := mock.ExpectExec(updateExtraSectionQuery).
				WithArgs(tt.section.Title, tt.section.Icon, tt.id)
			switch {
			case tt.execErr != nil:
				exp.WillReturnError(tt.execErr)
			case tt.resultErr != nil:
				exp.WillReturnResult(sqlmock.NewErrorResult(tt.resultErr))
			default:
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Update(context.Background(), tt.section, tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestExtraSectionRepository_Delete(t *testing.T) {
	tests := []struct {
		name              string
		id                int
		linksExecErr      error
		sectionExecErr    error
		sectionRows       int64
		commitErr         error
		err               error
	}{
		{
			name:        "extra section not found",
			id:          10,
			sectionRows: 0,
			err:         errs.ErrExtraSectionNotFound,
		},
		{
			name:           "delete links exec error",
			id:             10,
			linksExecErr:   errs.ErrDatabaseDown,
			err:            errs.ErrDatabaseDown,
		},
		{
			name:           "delete section exec error",
			id:             10,
			sectionExecErr: errs.ErrDatabaseDown,
			err:            errs.ErrDatabaseDown,
		},
		{
			name:        "accept a valid id",
			id:          10,
			sectionRows: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestExtraSectionRepo(t)
			mock.ExpectBegin()
			linksExp := mock.ExpectExec(deleteExtraLinksBySectionQuery).WithArgs(tt.id)
			if tt.linksExecErr != nil {
				linksExp.WillReturnError(tt.linksExecErr)
			} else {
				linksExp.WillReturnResult(sqlmock.NewResult(0, 1))
				sectionExp := mock.ExpectExec(deleteExtraSectionQuery).WithArgs(tt.id)
				if tt.sectionExecErr != nil {
					sectionExp.WillReturnError(tt.sectionExecErr)
				} else {
					sectionExp.WillReturnResult(sqlmock.NewResult(0, tt.sectionRows))
				}
			}
			if tt.err == nil {
				mock.ExpectCommit()
			}

			err := repo.Delete(context.Background(), tt.id)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}
