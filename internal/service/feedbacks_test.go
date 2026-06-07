package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeFeedbackRepo implements repository.FeedbackRepository for service tests.
type fakeFeedbackRepo struct {
	createCalls    int
	createFeedback models.Feedback
	createErr      error

	listCalls  int
	listLimit  int
	listOffset int
	listQ      string
	listStatus string
	listErr    error

	deleteCalls int
	deleteID    int
	deleteErr   error

	updateCalls  int
	updateID     int
	updateStatus string
	updateErr    error
}

func (f *fakeFeedbackRepo) Create(ctx context.Context, feedback models.Feedback) error {
	f.createCalls++
	f.createFeedback = feedback
	return f.createErr
}

func (f *fakeFeedbackRepo) List(ctx context.Context, limit, offset int, q, status string) ([]models.Feedback, error) {
	f.listCalls++
	f.listLimit = limit
	f.listOffset = offset
	f.listQ = q
	f.listStatus = status
	return nil, f.listErr
}

func (f *fakeFeedbackRepo) Update(ctx context.Context, status string, id int) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = id
	return f.updateErr
}

func (f *fakeFeedbackRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func TestFeedbackService_Create(t *testing.T) {
	tests := []struct {
		name         string
		feedback     models.Feedback
		createErr    error // simulate repo failure; nil for validation-only cases
		err          error // expected service error (use same value as createErr when wrapped)
		resultWanted *models.Feedback
	}{
		{
			name:         "rating shouldn't be 0",
			feedback:     models.Feedback{Category: "performance", Rating: 0, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackCategoryAndRatingRequired,
			resultWanted: nil,
		},
		{
			name:         "rating shouldn't be smaller than 0",
			feedback:     models.Feedback{Category: "performance", Rating: -10, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackInvalidRating,
			resultWanted: nil,
		},
		{
			name:         "rating shouldn't be greater than 6",
			feedback:     models.Feedback{Category: "performance", Rating: 6, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackInvalidRating,
			resultWanted: nil,
		},
		{
			name:         "accept a normal rating",
			feedback:     models.Feedback{Category: "performance", Rating: 4, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "performance", Rating: 4, Message: "nice"},
		},
		{
			name:         "reject empty category without space",
			feedback:     models.Feedback{Category: "", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackCategoryAndRatingRequired,
			resultWanted: nil,
		},
		{
			name:         "reject empty category with space",
			feedback:     models.Feedback{Category: "   ", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackCategoryAndRatingRequired,
			resultWanted: nil,
		},
		{
			name:         "reject invalid string category",
			feedback:     models.Feedback{Category: "hello", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackInvalidCategory,
			resultWanted: nil,
		},
		{
			name:         "reject invalid numerical category",
			feedback:     models.Feedback{Category: "123", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          errs.ErrFeedbackInvalidCategory,
			resultWanted: nil,
		},
		{
			name:         "accept valid category: ui/ux",
			feedback:     models.Feedback{Category: "ui/ux", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "ui/ux", Rating: 3, Message: "nice"},
		},
		{
			name:         "accept valid category: content",
			feedback:     models.Feedback{Category: "content", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "content", Rating: 3, Message: "nice"},
		},
		{
			name:         "accept valid category: functionality",
			feedback:     models.Feedback{Category: "functionality", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "functionality", Rating: 3, Message: "nice"},
		},
		{
			name:         "accept valid category: performance",
			feedback:     models.Feedback{Category: "performance", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "performance", Rating: 3, Message: "nice"},
		},
		{
			name:         "accept valid category: accessibility",
			feedback:     models.Feedback{Category: "accessibility", Rating: 3, Message: "nice"},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "accessibility", Rating: 3, Message: "nice"},
		},
		{
			name:         "repo create error",
			feedback:     models.Feedback{Category: "performance", Rating: 3, Message: "nice"},
			createErr:    errs.ErrDatabaseDown,
			err:          errs.ErrDatabaseDown,
			resultWanted: nil,
		},
		{
			name:         "trims and persists",
			feedback:     models.Feedback{Category: "    performance    ", Rating: 3, Message: "    nice   "},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Feedback{Category: "performance", Rating: 3, Message: "nice"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFeedbackRepo{createErr: tt.createErr}
			s := NewFeedbackService(repo)
			err := s.Create(context.Background(), tt.feedback)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.createErr != nil {
					wantCalls = 1
				}
				if repo.createCalls != wantCalls {
					t.Fatalf("repo.Create calls: got %d want %d", repo.createCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Create Feedback Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createFeedback, *tt.resultWanted) {
				t.Fatalf("repo.Create feedback: got %+v want %+v", repo.createFeedback, *tt.resultWanted)
			}
		})
	}
}

func TestFeedbackService_Update(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantStatus string
		idStr      string
		id         int
		err        error
		updateErr  error
	}{
		{
			name:      "reject non numerical id",
			idStr:     "abc",
			err:       errs.ErrFeedbackInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			status:    "read",
			err:       errs.ErrFeedbackInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			status:    "read",
			err:       errs.ErrFeedbackInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrFeedbackInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrFeedbackInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty status with spaces",
			idStr:     "10",
			status:    "  ",
			err:       errs.ErrStatusRequired,
			updateErr: nil,
		},
		{
			name:      "reject empty status without spaces",
			idStr:     "10",
			status:    "",
			err:       errs.ErrStatusRequired,
			updateErr: nil,
		},
		{
			name:      "reject invalid status",
			idStr:     "10",
			status:    "pending",
			err:       errs.ErrFeedbackInvalidStatus,
			updateErr: nil,
		},
		{
			name:      "repo update error",
			idStr:     "10",
			id:        10,
			status:    "read",
			err:       errs.ErrDatabaseDown,
			updateErr: errs.ErrDatabaseDown,
		},
		{
			name:       "accept valid read status",
			idStr:      "10",
			id:         10,
			status:     "read",
			wantStatus: "read",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept valid new status",
			idStr:      "10",
			id:         10,
			status:     "new",
			wantStatus: "new",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept a valid id and status with spaces",
			idStr:      "   10   ",
			id:         10,
			status:     "  read  ",
			wantStatus: "read",
			err:        nil,
			updateErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFeedbackRepo{updateErr: tt.updateErr}
			s := NewFeedbackService(repo)
			err := s.Update(context.Background(), tt.status, tt.idStr)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.updateErr != nil {
					wantCalls = 1
				}
				if repo.updateCalls != wantCalls {
					t.Fatalf("repo.Update calls: got %d want %d", repo.updateCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Update Feedback Service succeeded, want error %v", tt.err)
			}
			if repo.updateCalls != 1 {
				t.Fatalf("repo.Update calls: got %d want 1", repo.updateCalls)
			}
			if repo.updateID != tt.id || repo.updateStatus != tt.wantStatus {
				t.Fatalf("repo.Update args: got id=%d status=%q, want id=%d status=%q", repo.updateID, repo.updateStatus, tt.id, tt.wantStatus)
			}
		})
	}
}

func TestFeedbackService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		idStr     string
		id        int
		err       error
		deleteErr error
	}{
		{
			name:      "reject non numerical id",
			idStr:     "ABD",
			err:       errs.ErrFeedbackInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			err:       errs.ErrFeedbackInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			err:       errs.ErrFeedbackInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrFeedbackInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrFeedbackInvalidID,
			deleteErr: nil,
		},
		{
			name:      "repo delete error",
			idStr:     "10",
			err:       errs.ErrDatabaseDown,
			deleteErr: errs.ErrDatabaseDown,
		},
		{
			name:      "accept a valid id",
			idStr:     "10",
			id:        10,
			err:       nil,
			deleteErr: nil,
		},
		{
			name:      "accept a valid id with spaces",
			idStr:     "  10  ",
			id:        10,
			err:       nil,
			deleteErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFeedbackRepo{deleteErr: tt.deleteErr}
			s := NewFeedbackService(repo)
			err := s.Delete(context.Background(), tt.idStr)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.deleteErr != nil {
					wantCalls = 1
				}
				if repo.deleteCalls != wantCalls {
					t.Fatalf("repo.Delete calls: got %d want %d", repo.deleteCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Delete Feedback Service succeeded, want error %v", tt.err)
			}
			if repo.deleteCalls != 1 {
				t.Fatalf("repo.Delete calls: got %d want 1", repo.deleteCalls)
			}
			if repo.deleteID != tt.id {
				t.Fatalf("repo.DeleteID : got %v, want %v", repo.deleteID, tt.id)
			}
		})
	}
}

func TestFeedbackService_List(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		q       string
		status  string
		err     error
		listErr error
	}{
		{
			name:    "reject offset < 0",
			limit:   10,
			offset:  -10,
			q:       "",
			status:  "read",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit > 100",
			limit:   110,
			offset:  10,
			q:       "",
			status:  "read",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "accept limit = 100",
			limit:   100,
			offset:  10,
			q:       "",
			status:  "new",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "reject limit = 0",
			limit:   0,
			offset:  10,
			q:       "",
			status:  "new",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit < 0",
			limit:   -10,
			offset:  10,
			q:       "",
			status:  "new",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "accept empty status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "reject invalid status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     errs.ErrFeedbackInvalidStatus,
			listErr: nil,
		},
		{
			name:    "accept new status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "new",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "accept a q",
			limit:   10,
			offset:  10,
			q:       "Performance",
			status:  "new",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "repo list error",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "new",
			err:     errs.ErrDatabaseDown,
			listErr: errs.ErrDatabaseDown,
		},
		{
			name:    "accept read status",
			limit:   50,
			offset:  100,
			q:       "",
			status:  "read",
			err:     nil,
			listErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFeedbackRepo{listErr: tt.listErr}
			s := NewFeedbackService(repo)
			_, err := s.List(context.Background(), tt.limit, tt.offset, tt.q, tt.status)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.listErr != nil {
					wantCalls = 1
				}
				if repo.listCalls != wantCalls {
					t.Fatalf("repo.List calls: got %d want %d", repo.listCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("List Feedback Service succeeded, want error %v", tt.err)
			}
			if repo.listCalls != 1 {
				t.Fatalf("repo.List calls: got %d want 1", repo.listCalls)
			}
			if repo.listLimit != tt.limit || repo.listOffset != tt.offset || repo.listQ != tt.q || repo.listStatus != tt.status {
				t.Fatalf("repo.List args: got limit=%d offset=%d q=%q status=%q want limit=%d offset=%d q=%q status=%q",
					repo.listLimit, repo.listOffset, repo.listQ, repo.listStatus,
					tt.limit, tt.offset, tt.q, tt.status)
			}
		})
	}
}
