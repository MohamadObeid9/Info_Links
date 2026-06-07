package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"infolinks-backend/internal/errs"
)

func newTestContentRepo(t *testing.T) (ContentRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresContentRepository(db), mock
}

func TestContentRepository_Get(t *testing.T) {
	sampleJSON := `{"programs":[],"years":[],"semesters":[],"courses":[],"links":[],"extra_sections":[],"extra_links":[]}`

	tests := []struct {
		name     string
		queryErr error
		err      error
	}{
		{
			name: "returns content json",
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestContentRepo(t)
			exp := mock.ExpectQuery(getContentQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(sqlmock.NewRows([]string{"json"}).AddRow(sampleJSON))
			}

			got, err := repo.Get(context.Background())
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
				t.Fatalf("Get: %v", err)
			}
			if !reflect.DeepEqual(got, []byte(sampleJSON)) {
				t.Fatalf("got %q, want %q", got, sampleJSON)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}
