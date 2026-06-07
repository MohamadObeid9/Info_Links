package repository

import (
	"context"
	"errors"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestLinkRepo(t *testing.T) (LinkRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresLinkRepository(db), mock
}

func TestLinkRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		link    models.Link
		execErr error
		err     error
	}{
		{
			name: "insert link",
			link: models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 5},
		},
		{
			name:    "insert exec error",
			link:    models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 5},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestLinkRepo(t)
			exp := mock.ExpectExec(insertLinkQuery).
				WithArgs(tt.link.CourseID, tt.link.Type, tt.link.URL, tt.link.Label, tt.link.Note, tt.link.ContentType, tt.link.DisplayOrder)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.link)
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

func TestLinkRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "link not found",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrLinkNotFound,
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
			repo, mock := newTestLinkRepo(t)
			exp := mock.ExpectExec(deleteLinkQuery).WithArgs(tt.id)
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

func TestLinkRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		link         models.Link
		id           int
		execErr      error
		resultErr    error
		rowsAffected int64
		err          error
	}{
		{
			name:         "link not found",
			id:           10,
			link:         models.Link{Type: "telegram", URL: "http://fake.test", Label: "link 1", Note: "", ContentType: nil},
			rowsAffected: 0,
			err:          errs.ErrLinkNotFound,
		},
		{
			name:    "update exec error",
			id:      10,
			link:    models.Link{Type: "telegram", URL: "http://fake.test", Label: "link 1", Note: "", ContentType: nil},
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:      "rows affected error",
			id:        10,
			link:      models.Link{Type: "telegram", URL: "http://fake.test", Label: "link 1", Note: "", ContentType: nil},
			resultErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid link",
			id:           10,
			link:         models.Link{Type: "telegram", URL: "http://fake.test", Label: "link 1", Note: "", ContentType: nil},
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestLinkRepo(t)
			exp := mock.ExpectExec(updateLinkQuery).
				WithArgs(tt.link.Type, tt.link.URL, tt.link.Label, tt.link.Note, tt.link.ContentType, tt.id)
			switch {
			case tt.execErr != nil:
				exp.WillReturnError(tt.execErr)
			case tt.resultErr != nil:
				exp.WillReturnResult(sqlmock.NewErrorResult(tt.resultErr))
			default:
				exp.WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))
			}

			err := repo.Update(context.Background(), tt.link, tt.id)
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}
