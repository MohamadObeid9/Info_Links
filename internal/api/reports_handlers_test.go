package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeReportService implements reportService for handler tests.
type fakeReportService struct {
	createCalls  int
	createReport models.Report
	createErr    error

	listCalls  int
	listLimit  int
	listOffset int
	listQ      string
	listStatus string
	listResult []models.Report
	listErr    error

	deleteCalls int
	deleteID    string
	deleteErr   error

	updateCalls  int
	updateStatus string
	updateID     string
	updateErr    error
}

func (f *fakeReportService) Create(ctx context.Context, report models.Report) error {
	f.createCalls++
	f.createReport = report
	return f.createErr
}

func (f *fakeReportService) List(ctx context.Context, limit, offset int, q, status string) ([]models.Report, error) {
	f.listCalls++
	f.listLimit = limit
	f.listOffset = offset
	f.listQ = q
	f.listStatus = status
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeReportService) Delete(ctx context.Context, id string) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeReportService) Update(ctx context.Context, status, idStr string) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = idStr
	return f.updateErr
}

func testReportsHandler(t *testing.T, svc *fakeReportService) *Handler {
	t.Helper()
	h, err := NewHandler(Dependencies{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:     []byte("test-jwt-secret"),
		ReportService: svc,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

func TestHandlePostReport(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		reportWanted *models.Report
	}{
		{
			name:         "201 when service accepts the report",
			body:         `{"course_name":"A","link_url":"https://example.com","description":""}`,
			createErr:    nil,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			reportWanted: &models.Report{CourseName: "A", LinkURL: "https://example.com", Description: ""},
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 when service returns validation error",
			body:         `{"course_name":"x","link_url":"https://x","description":""}`,
			createErr:    errs.ErrCourseNameAndLinkUrlAreRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course name and link URL are required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"course_name":"A","link_url":"https://example.com","description":""}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeReportService{createErr: tt.createErr}
			h := testReportsHandler(t, fake)
			req := httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handlePostReport(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusCreated {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.createCalls != tt.wantCalls {
					t.Fatalf("service.Create calls: got %d want %d", fake.createCalls, tt.wantCalls)
				}
				return
			}

			// from now on , only success cases pass
			if fake.createCalls != tt.wantCalls {
				t.Fatalf("service.Create calls: got %d want %d", fake.createCalls, tt.wantCalls)
			}
			if tt.reportWanted == nil {
				t.Fatal("success case must set reportWanted")
			}
			if !reflect.DeepEqual(fake.createReport, *tt.reportWanted) {
				t.Fatalf("service.Create report: got %+v want %+v", fake.createReport, *tt.reportWanted)
			}
			if rr.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", rr.Body.String())
			}
		})
	}
}

func TestHandleAdminGetReports(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		listErr      error
		listResult   []models.Report
		statusWanted int
		errMsg       string
		wantLimit    int
		wantOffset   int
		wantQ        string
		wantStatus   string
	}{
		{
			name:         "reject empty status",
			target:       "/api/admin/reports?limit=10&offset=10",
			listErr:      errs.ErrReportStatusRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status is required",
		},
		{
			name:         "reject invalid status",
			target:       "/api/admin/reports?limit=10&offset=10&status=pending",
			listErr:      errs.ErrInvalidReportStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be open or resolved",
		},
		{
			name:         "reject list invalid params",
			target:       "/api/admin/reports?limit=10&offset=10&status=open",
			listErr:      errs.ErrListReportInvalidParams,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Limit should be between 1-100 and Offset >= 0",
		},
		{
			name:         "500 when service fails",
			target:       "/api/admin/reports?limit=10&offset=10&status=open",
			listErr:      errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
		},
		{
			name:         "accept open status",
			target:       "/api/admin/reports?limit=10&offset=10&status=open",
			listResult:   []models.Report{{ID: 1, CourseName: "Linux", Status: "open"}},
			statusWanted: http.StatusOK,
			wantLimit:    10,
			wantOffset:   10,
			wantStatus:   "open",
		},
		{
			name:         "accept resolved status with search",
			target:       "/api/admin/reports?limit=50&offset=100&q=Linux&status=resolved",
			listResult:   []models.Report{{ID: 2, CourseName: "Go", Status: "resolved"}},
			statusWanted: http.StatusOK,
			wantLimit:    50,
			wantOffset:   100,
			wantQ:        "Linux",
			wantStatus:   "resolved",
		},
		{
			name:         "default pagination and trimmed status",
			target:       "/api/admin/reports?status=+open+",
			listResult:   []models.Report{},
			statusWanted: http.StatusOK,
			wantLimit:    25,
			wantOffset:   0,
			wantStatus:   "open",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeReportService{listErr: tt.listErr, listResult: tt.listResult}
			h := testReportsHandler(t, fake)
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetReports(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.listCalls != 1 {
					t.Fatalf("service.List calls: got %d want 1", fake.listCalls)
				}
				return
			}

			if fake.listCalls != 1 {
				t.Fatalf("service.List calls: got %d want 1", fake.listCalls)
			}
			if fake.listLimit != tt.wantLimit || fake.listOffset != tt.wantOffset || fake.listQ != tt.wantQ || fake.listStatus != tt.wantStatus {
				t.Fatalf("service.List args: got limit=%d offset=%d q=%q status=%q want limit=%d offset=%d q=%q status=%q",
					fake.listLimit, fake.listOffset, fake.listQ, fake.listStatus,
					tt.wantLimit, tt.wantOffset, tt.wantQ, tt.wantStatus)
			}
			var got []models.Report
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}

func TestHandleAdminUpdateReport(t *testing.T) {
	tests := []struct {
		name         string
		pathID       string
		body         string
		updateErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
		wantStatus   string
	}{
		{
			name:         "400 invalid JSON body",
			pathID:       "10",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 invalid report id",
			pathID:       "abc",
			body:         `{"status":"open"}`,
			updateErr:    errs.ErrInvalidReportID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid report id",
			wantCalls:    1,
		},
		{
			name:         "404 report not found",
			pathID:       "10",
			body:         `{"status":"open"}`,
			updateErr:    errs.ErrReportNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Report not found",
			wantCalls:    1,
		},
		{
			name:         "400 status is required",
			pathID:       "10",
			body:         `{"status":""}`,
			updateErr:    errs.ErrReportStatusRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status is required",
			wantCalls:    1,
		},
		{
			name:         "400 invalid status",
			pathID:       "10",
			body:         `{"status":"pending"}`,
			updateErr:    errs.ErrInvalidReportStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be open or resolved",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"status":"open"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept valid open status",
			pathID:       "10",
			body:         `{"status":"open"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "open",
		},
		{
			name:         "accept valid resolved status",
			pathID:       "10",
			body:         `{"status":"resolved"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "resolved",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeReportService{updateErr: tt.updateErr}
			h := testReportsHandler(t, fake)
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/reports/"+tt.pathID, bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminUpdateReport(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.updateCalls != tt.wantCalls {
					t.Fatalf("service.Update calls: got %d want %d", fake.updateCalls, tt.wantCalls)
				}
				return
			}

			if fake.updateCalls != tt.wantCalls {
				t.Fatalf("service.Update calls: got %d want %d", fake.updateCalls, tt.wantCalls)
			}
			if fake.updateID != tt.wantID || fake.updateStatus != tt.wantStatus {
				t.Fatalf("service.Update args: got id=%q status=%q want id=%q status=%q",
					fake.updateID, fake.updateStatus, tt.wantID, tt.wantStatus)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("response status: got %q want ok", got["status"])
			}
		})
	}
}

func TestHandleAdminDeleteReport(t *testing.T) {
	tests := []struct {
		name         string
		pathID       string
		deleteErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
	}{
		{
			name:         "400 invalid report id",
			pathID:       "abc",
			deleteErr:    errs.ErrInvalidReportID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid report id",
			wantCalls:    1,
		},
		{
			name:         "404 report not found",
			pathID:       "10",
			deleteErr:    errs.ErrReportNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Report not found",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			deleteErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept a valid id",
			pathID:       "10",
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
		},
		{
			name:         "accept a valid id with spaces",
			pathID:       "  10  ",
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "  10  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeReportService{deleteErr: tt.deleteErr}
			h := testReportsHandler(t, fake)
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/reports/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteReport(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.deleteCalls != tt.wantCalls {
					t.Fatalf("service.Delete calls: got %d want %d", fake.deleteCalls, tt.wantCalls)
				}
				return
			}

			if fake.deleteCalls != tt.wantCalls {
				t.Fatalf("service.Delete calls: got %d want %d", fake.deleteCalls, tt.wantCalls)
			}
			if fake.deleteID != tt.wantID {
				t.Fatalf("service.Delete id: got %q want %q", fake.deleteID, tt.wantID)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("response status: got %q want ok", got["status"])
			}
		})
	}
}
