package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeReportRepo implements repository.ReportRepository for service tests.
type fakeReportRepo struct {
	createCalls  int
	createReport models.Report
	createErr    error

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

func (f *fakeReportRepo) Create(ctx context.Context, report models.Report) error {
	f.createCalls++
	f.createReport = report
	return f.createErr
}

func (f *fakeReportRepo) List(ctx context.Context, limit, offset int, q, status string) ([]models.Report, error) {
	f.listCalls++
	f.listLimit = limit
	f.listOffset = offset
	f.listQ = q
	f.listStatus = status
	return nil, f.listErr
}

func (f *fakeReportRepo) Update(ctx context.Context, status string, id int) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = id
	return f.updateErr
}

func (f *fakeReportRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func TestReportService_List(t *testing.T) {
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
			status:  "open",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit > 100",
			limit:   110,
			offset:  10,
			q:       "",
			status:  "open",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "accept limit = 100",
			limit:   100,
			offset:  10,
			q:       "",
			status:  "open",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "reject limit = 0",
			limit:   0,
			offset:  10,
			q:       "",
			status:  "open",
			err:     errs.ErrInvalidParams,
			listErr: nil,
		},
		{
			name:    "reject limit < 0",
			limit:   -10,
			offset:  10,
			q:       "",
			status:  "open",
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
			err:     errs.ErrReportInvalidStatus,
			listErr: nil,
		},
		{
			name:    "accept open status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "open",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "accept a q",
			limit:   10,
			offset:  10,
			q:       "Linux",
			status:  "open",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "repo list error",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "open",
			err:     errs.ErrDatabaseDown,
			listErr: errs.ErrDatabaseDown,
		},
		{
			name:    "accept resolved status",
			limit:   50,
			offset:  100,
			q:       "",
			status:  "resolved",
			err:     nil,
			listErr: nil,
		},
		{
			name:    "accept rejected status",
			limit:   10,
			offset:  10,
			q:       "",
			status:  "rejected",
			err:     nil,
			listErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeReportRepo{listErr: tt.listErr}
			s := NewReportService(repo)
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
				t.Fatalf("List Report Service succeeded, want error %v", tt.err)
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

func TestReportService_Create(t *testing.T) {
	tests := []struct {
		name         string
		report       models.Report
		createErr    error // simulate repo failure; nil for validation-only cases
		err          error // expected service error (use same value as createErr when wrapped)
		resultWanted *models.Report
	}{
		{
			name:         "course name is required",
			report:       models.Report{CourseName: "", LinkURL: "http://fake.link", Description: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "link url is required",
			report:       models.Report{CourseName: "fake course", LinkURL: "", Description: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "course name shouldn't be empty",
			report:       models.Report{CourseName: "  ", LinkURL: "http://fake.link", Description: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "link url shouldn't be empty",
			report:       models.Report{CourseName: "fake course", LinkURL: "  ", Description: ""},
			createErr:    nil,
			err:          errs.ErrCourseNameAndLinkUrlRequired,
			resultWanted: nil,
		},
		{
			name:         "repo create error",
			report:       models.Report{CourseName: "My Course", LinkURL: "https://test.com", Description: "note"},
			createErr:    errs.ErrDatabaseDown,
			err:          errs.ErrDatabaseDown,
			resultWanted: nil,
		},
		{
			name:         "trims and persists",
			report:       models.Report{CourseName: "  My Course  ", LinkURL: "  https://test.com  ", Description: "  note  "},
			createErr:    nil,
			err:          nil,
			resultWanted: &models.Report{CourseName: "My Course", LinkURL: "https://test.com", Description: "note"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeReportRepo{createErr: tt.createErr}
			s := NewReportService(repo)
			err := s.Create(context.Background(), tt.report)
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
				t.Fatalf("Create Report Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createReport, *tt.resultWanted) {
				t.Fatalf("repo.Create report: got %+v want %+v", repo.createReport, *tt.resultWanted)
			}
		})
	}
}

func TestReportService_Delete(t *testing.T) {
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
			err:       errs.ErrReportInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			err:       errs.ErrReportInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			err:       errs.ErrReportInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrReportInvalidID,
			deleteErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrReportInvalidID,
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
			repo := &fakeReportRepo{deleteErr: tt.deleteErr}
			s := NewReportService(repo)
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
				t.Fatalf("Delete Report Service succeeded, want error %v", tt.err)
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

func TestReportService_Update(t *testing.T) {
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
			err:       errs.ErrReportInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id with spaces",
			idStr:     "  ",
			status:    "open",
			err:       errs.ErrReportInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject empty id without spaces",
			idStr:     "",
			status:    "open",
			err:       errs.ErrReportInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id = 0",
			idStr:     "0",
			err:       errs.ErrReportInvalidID,
			updateErr: nil,
		},
		{
			name:      "reject id < 0",
			idStr:     "-10",
			err:       errs.ErrReportInvalidID,
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
			err:       errs.ErrReportInvalidStatus,
			updateErr: nil,
		},
		{
			name:      "repo update error",
			idStr:     "10",
			id:        10,
			status:    "open",
			err:       errs.ErrDatabaseDown,
			updateErr: errs.ErrDatabaseDown,
		},
		{
			name:       "accept valid open status",
			idStr:      "10",
			id:         10,
			status:     "open",
			wantStatus: "open",
			err:        nil,
			updateErr:  nil,
		},
		{
			name:       "accept valid resolved status",
			idStr:      "10",
			id:         10,
			status:     "resolved",
			wantStatus: "resolved",
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
			status:     "  open  ",
			wantStatus: "open",
			err:        nil,
			updateErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeReportRepo{updateErr: tt.updateErr}
			s := NewReportService(repo)
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
				t.Fatalf("Update Report Service succeeded, want error %v", tt.err)
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
