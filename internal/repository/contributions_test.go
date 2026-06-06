package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func newTestContributionRepo(t *testing.T) (ContributionRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresContributionRepository(db), mock
}

func TestPostgresContributionRepository_Create(t *testing.T) {
	tests := []struct {
		name         string
		contribution models.Contribution
		execErr      error
		err          error
	}{
		{
			name:         "insert contribution",
			contribution: models.Contribution{CourseName: "My Course", LinkURL: "https://test.com", Note: "note"},
		},
		{
			name:         "insert exec error",
			contribution: models.Contribution{CourseName: "My Course", LinkURL: "https://test.com", Note: "note"},
			execErr:      errs.ErrDatabaseDown,
			err:          errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestContributionRepo(t)
			exp := mock.ExpectExec(insertContributionQuery).
				WithArgs(tt.contribution.CourseName, tt.contribution.LinkURL, tt.contribution.Note)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.contribution)
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

func TestPostgresContributionRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "contribution not found",
			id:           909,
			rowsAffected: 0,
			err:          errs.ErrContributionNotFound,
		},
		{
			name:    "delete exec error",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept a valid id",
			id:           15,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestContributionRepo(t)
			exp := mock.ExpectExec(deleteContributionQuery).WithArgs(tt.id)
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

func TestPostgresContributionRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "Contribution not found",
			status:       "pending",
			id:           909,
			rowsAffected: 0,
			err:          errs.ErrContributionNotFound,
		},
		{
			name:    "update exec error",
			status:  "approved",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid pending status",
			status:       "pending",
			id:           15,
			rowsAffected: 1,
		},
		{
			name:         "accept valid approved status",
			status:       "approved",
			id:           15,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestContributionRepo(t)
			exp := mock.ExpectExec(updateContributionQuery).WithArgs(tt.status, tt.id)
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

func TestPostgresContributionRepository_List(t *testing.T) {
	sampleRow := models.Contribution{
		ID:         3,
		CourseName: "Linux",
		LinkURL:    "https://example.com",
		Note:       "broken link",
		Status:     "pending",
		CreatedAt:  "2026-03-15T11:26:45Z",
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
		wantResult []models.Contribution
	}{
		{
			name:       "list without filters",
			limit:      35,
			offset:     0,
			query:      listContributionsNoFilterQuery,
			queryArgs:  []any{35, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Note, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Contribution{sampleRow},
		},
		{
			name:       "list with search query",
			limit:      15,
			offset:     25,
			q:          "Linux",
			status:     "pending",
			query:      listContributionsWithQStatusQuery,
			queryArgs:  []any{"%Linux%", "pending", 15, 25},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Note, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Contribution{sampleRow},
		},
		{
			name:       "list with status only",
			limit:      10,
			offset:     10,
			status:     "approved",
			query:      listContributionsWithStatusQuery,
			queryArgs:  []any{"approved", 10, 10},
			rows:       [][]any{{2, "Go", "https://go.dev", "", "approved", "2024-02-01T00:00:00Z"}},
			wantResult: []models.Contribution{{ID: 2, CourseName: "Go", LinkURL: "https://go.dev", Status: "approved", CreatedAt: "2024-02-01T00:00:00Z"}},
		},
		{
			name:       "list with q only",
			limit:      10,
			offset:     0,
			q:          "Linux",
			query:      listContributionsWithQQuery,
			queryArgs:  []any{"%Linux%", 10, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Note, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Contribution{sampleRow},
		},
		{
			name:      "list query error",
			limit:     10,
			offset:    0,
			status:    "approved",
			query:     listContributionsWithStatusQuery,
			queryArgs: []any{"approved", 10, 0},
			queryErr:  errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:      "list rows error",
			limit:     10,
			offset:    0,
			status:    "pending",
			query:     listContributionsWithStatusQuery,
			queryArgs: []any{"pending", 10, 0},
			rows:      [][]any{{sampleRow.ID, sampleRow.CourseName, sampleRow.LinkURL, sampleRow.Note, sampleRow.Status, sampleRow.CreatedAt}},
			rowsErr:   errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestContributionRepo(t)
			exp := mock.ExpectQuery(tt.query).WithArgs(driverValues(tt.queryArgs)...)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				cols := []string{"id", "course_name", "link_url", "note", "status", "created_at"}
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
