package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeContributionRepo implements repository.ContributionRepository for service tests.
type fakeContributionRepo struct {
	createCalls        int
	createContribution models.Contribution
	createErr          error

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

func (f *fakeContributionRepo) Create(ctx context.Context, contribute models.Contribution) error {
	f.createCalls++
	f.createContribution = contribute
	return f.createErr
}

func (f *fakeContributionRepo) List(ctx context.Context, limit, offset int, q, status string) ([]models.Contribution, error) {
	f.listCalls++
	f.listLimit = limit
	f.listOffset = offset
	f.listQ = q
	f.listStatus = status
	return nil, f.listErr
}

func (f *fakeContributionRepo) Update(ctx context.Context, status string, id int) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = id
	return f.updateErr
}

func (f *fakeContributionRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func TestContributionService_List(t *testing.T) {
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
			status:  "pending",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit > 100",
			limit:   110,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "accept limit = 100",
			limit:   100,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "reject limit = 0",
			limit:   0,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit < 0",
			limit:   -10,
			offset:  10,
			q:       "",
			status:  "pending",
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
			status:  "open",
			err:     errs.ErrContributionInvalidStatus,
			listErr: nil,
		},
		{
			name:    "accept pending status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "accept a q",
			limit:   10,
			offset:  10,
			q:       "Linux",
			status:  "pending",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "repo list error",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "pending",
			err:     errs.ErrDatabaseDown,
			listErr: errs.ErrDatabaseDown,
		},
		{
			name:    "accept approved status",
			limit:   50,
			offset:  100,
			q:       "",
			status:  "approved",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "accept rejected status",
			limit:   10,
			offset:  0,
			q:       "",
			status:  "rejected",
			err:     nil,
			listErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeContributionRepo{listErr: tt.listErr}
			s := NewContributionService(repo)
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
				t.Fatalf("List Contribution Service succeeded, want error %v", tt.err)
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

func TestContributionService_Create(t *testing.T) {
	tests := []struct {
		name         string
		contribution models.Contribution
		createErr    error // simulate repo failure; nil for validation-only cases
		err          error // expected service error (use same value as createErr when wrapped)
		resultWanted *models.Contribution
	}{
		{
			name:         "course name is required",
			contribution: models.Contribution{CourseName: "", LinkURL: "http://fake.link", Note: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "link url is required",
			contribution: models.Contribution{CourseName: "fake course", LinkURL: "", Note: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "course name shouldn't be empty",
			contribution: models.Contribution{CourseName: "  ", LinkURL: "http://fake.link", Note: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "link url shouldn't be empty",
			contribution: models.Contribution{CourseName: "fake course", LinkURL: "  ", Note: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "repo create error",
			contribution: models.Contribution{CourseName: "My Course", LinkURL: "https://test.com", Note: "note"},
			createErr:    errs.ErrDatabaseDown,
			err:          errs.ErrDatabaseDown,
			resultWanted: nil,
		},
		{
			name:         "trims and persists",
			contribution: models.Contribution{CourseName: "  My Course  ", LinkURL: "  https://test.com  ", Note: "  note  "},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Contribution{CourseName: "My Course", LinkURL: "https://test.com", Note: "note"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeContributionRepo{createErr: tt.createErr}
			s := NewContributionService(repo)
			err := s.Create(context.Background(), tt.contribution)
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
				t.Fatalf("Create Contribution Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createContribution, *tt.resultWanted) {
				t.Fatalf("repo.Create contribution: got %+v want %+v", repo.createContribution, *tt.resultWanted)
			}
		})
	}
}

func TestContributionService_Delete(t *testing.T) {
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
			err:       errs.ErrContributionInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			err:       errs.ErrContributionInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			err:       errs.ErrContributionInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrContributionInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrContributionInvalidID,
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
			repo := &fakeContributionRepo{deleteErr: tt.deleteErr}
			s := NewContributionService(repo)
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
				t.Fatalf("Delete Contribution Service succeeded, want error %v", tt.err)
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

func TestContributionService_Update(t *testing.T) {
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
			err:       errs.ErrContributionInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			status:    "pending",
			err:       errs.ErrContributionInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			status:    "open",
			err:       errs.ErrContributionInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrContributionInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrContributionInvalidID,
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
			status:    "open",
			err:       errs.ErrContributionInvalidStatus,
			updateErr: nil,
		},
		{
			name:      "repo update error",
			idStr:     "10",
			id:        10,
			status:    "pending",
			err:       errs.ErrDatabaseDown,
			updateErr: errs.ErrDatabaseDown,
		},
		{
			name:       "accept valid pending status",
			idStr:      "10",
			id:         10,
			status:     "pending",
			wantStatus: "pending",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept valid approved status",
			idStr:      "10",
			id:         10,
			status:     "approved",
			wantStatus: "approved",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept valid rejected status",
			idStr:      "10",
			id:         10,
			status:     "rejected",
			wantStatus: "rejected",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept a valid id and status with spaces",
			idStr:      "   10   ",
			id:         10,
			status:     "  pending  ",
			wantStatus: "pending",
			err:        nil,
			updateErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeContributionRepo{updateErr: tt.updateErr}
			s := NewContributionService(repo)
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
				t.Fatalf("Update Contribution Service succeeded, want error %v", tt.err)
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
