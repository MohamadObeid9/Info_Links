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

func newTestFeedbackRepo(t *testing.T) (FeedbackRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresFeedbackRepository(db), mock
}

func TestFeedbackRepository_Create(t *testing.T) {
	tests := []struct {
		name     string
		feedback models.Feedback
		execErr  error
		err      error
	}{
		{
			name:     "insert feedback",
			feedback: models.Feedback{Category: "Performance", Rating: 5, Message: "very good resources"},
		},
		{
			name:     "insert exec error",
			feedback: models.Feedback{Category: "Performance", Rating: 5, Message: "very good resources"},
			execErr:  errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestFeedbackRepo(t)
			exp := mock.ExpectExec(insertFeedbackQuery).
				WithArgs(tt.feedback.Category, tt.feedback.Rating, tt.feedback.Message)
			if tt.execErr != nil {
				exp.WillReturnError(tt.execErr)
			} else {
				exp.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.Create(context.Background(), tt.feedback)
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

func TestFeedbackRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "feedback not found",
			status:       "read",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrFeedbackNotFound,
		},
		{
			name:    "update exec error",
			status:  "read",
			id:      10,
			execErr: errs.ErrDatabaseDown,
			err:     errs.ErrDatabaseDown,
		},
		{
			name:         "accept valid read status",
			status:       "read",
			id:           10,
			rowsAffected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestFeedbackRepo(t)
			exp := mock.ExpectExec(updateFeedbackQuery).WithArgs(tt.status, tt.id)
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

func TestFeedbackRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		execErr      error
		rowsAffected int64
		err          error
	}{
		{
			name:         "feedback not found",
			id:           99,
			rowsAffected: 0,
			err:          errs.ErrFeedbackNotFound,
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
			repo, mock := newTestFeedbackRepo(t)
			exp := mock.ExpectExec(deleteFeedbackQuery).WithArgs(tt.id)
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

func TestFeedbackRepository_List(t *testing.T) {
	sampleRow := models.Feedback{
		ID:        1,
		Category:  "Performance",
		Rating:    5,
		Message:   "very nice",
		Status:    "new",
		CreatedAt: "2026-01-01T00:00:00Z",
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
		wantResult []models.Feedback
	}{
		{
			name:       "list without filters",
			limit:      25,
			offset:     0,
			query:      listFeedbackNoFilterQuery,
			queryArgs:  []any{25, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.Category, sampleRow.Rating, sampleRow.Message, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Feedback{sampleRow},
		},
		{
			name:       "list with search query",
			limit:      10,
			offset:     5,
			q:          "Performance",
			status:     "new",
			query:      listFeedbackWithQStatusQuery,
			queryArgs:  []any{"%Performance%", "new", 10, 5},
			rows:       [][]any{{sampleRow.ID, sampleRow.Category, sampleRow.Rating, sampleRow.Message, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Feedback{sampleRow},
		},
		{
			name:       "list with status only",
			limit:      10,
			offset:     10,
			status:     "new",
			query:      listFeedbackWithStatusQuery,
			queryArgs:  []any{"new", 10, 10},
			rows:       [][]any{{2, "Performance", 5, "nice", "new", "2024-02-01T00:00:00Z"}},
			wantResult: []models.Feedback{{ID: 2, Category: "Performance", Rating: 5, Message: "nice", Status: "new", CreatedAt: "2024-02-01T00:00:00Z"}},
		},
		{
			name:       "list with q only",
			limit:      10,
			offset:     0,
			q:          "Performance",
			query:      listFeedbackWithQQuery,
			queryArgs:  []any{"%Performance%", 10, 0},
			rows:       [][]any{{sampleRow.ID, sampleRow.Category, sampleRow.Rating, sampleRow.Message, sampleRow.Status, sampleRow.CreatedAt}},
			wantResult: []models.Feedback{sampleRow},
		},
		{
			name:      "list query error",
			limit:     10,
			offset:    0,
			status:    "new",
			query:     listFeedbackWithStatusQuery,
			queryArgs: []any{"new", 10, 0},
			queryErr:  errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:      "list rows error",
			limit:     10,
			offset:    0,
			status:    "new",
			query:     listFeedbackWithStatusQuery,
			queryArgs: []any{"new", 10, 0},
			rows:      [][]any{{sampleRow.ID, sampleRow.Category, sampleRow.Rating, sampleRow.Message, sampleRow.Status, sampleRow.CreatedAt}},
			rowsErr:   errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestFeedbackRepo(t)
			exp := mock.ExpectQuery(tt.query).WithArgs(driverValues(tt.queryArgs)...)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				cols := []string{"id", "category", "rating", "message", "status", "created_at"}
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
