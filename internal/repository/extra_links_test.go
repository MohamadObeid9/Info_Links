package repository

import (
	"context"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestExtraLinkRepo(t *testing.T) (ExtraLinkRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresExtraLinkRepository(db), mock
}

func sectionID(v int) *int { return &v }

func TestExtraLinkRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		link    models.ExtraLink
		execErr error
		err     error
	}{
		{
			name: "insert extra link",
			link: models.ExtraLink{SectionID: sectionID(1), Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 0},
		},
		{
			name:    "insert exec error",
			link:    models.ExtraLink{SectionID: sectionID(1), Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 0},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestExtraLinkRepo(t)
			exp := mock.ExpectExec(insertExtraLinkQuery).
				WithArgs(*tt.link.SectionID, tt.link.Type, tt.link.URL, tt.link.Label, tt.link.Note, tt.link.ContentType, tt.link.DisplayOrder)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.link)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestExtraLinkRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		link         models.ExtraLink
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "extra link not found",
			id:           10,
			link:         models.ExtraLink{Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil},
			rowsAffected: 0,
			err:          errs.ErrExtraLinkNotFound,
		},
		{
			name:    "update exec error",
			id:      10,
			link:    models.ExtraLink{Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid link",
			id:           10,
			link:         models.ExtraLink{Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil},
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestExtraLinkRepo(t)
			exp := mock.ExpectExec(updateExtraLinkQuery).
				WithArgs(tt.link.Type, tt.link.URL, tt.link.Label, tt.link.Note, tt.link.ContentType, tt.id)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Update(context.Background(), tt.link, tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestExtraLinkRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "extra link not found",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrExtraLinkNotFound,
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
			repo, mock := newTestExtraLinkRepo(t)
			exp := mock.ExpectExec(deleteExtraLinkQuery).WithArgs(tt.id)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Delete(context.Background(), tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestExtraLinkRepository_List(t *testing.T) {
	columns := []string{"id", "section_id", "type", "url", "label", "note", "content_type", "display_order"}
	section := 1

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     []models.ExtraLink
		err      error
	}{
		{
			name: "returns extra links",
			rows: sqlmock.NewRows(columns).AddRow(1, section, "telegram", "https://fake.test", "link 1", "", nil, 0),
			want: []models.ExtraLink{{ID: 1, SectionID: &section, Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil, DisplayOrder: 0}},
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
			repo, mock := newTestExtraLinkRepo(t)
			exp := mock.ExpectQuery(listExtraLinksQuery)
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
