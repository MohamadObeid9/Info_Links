package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeContributionService implements ContributionService for handler tests.
type fakeContributionService struct {
	createCalls        int
	createContribution models.Contribution
	createErr          error

	listCalls  int
	listLimit  int
	listOffset int
	listQ      string
	listStatus string
	listResult []models.Contribution
	listErr    error

	deleteCalls int
	deleteID    string
	deleteErr   error

	updateCalls  int
	updateStatus string
	updateID     string
	updateErr    error
}

func (f *fakeContributionService) Create(ctx context.Context, contribution models.Contribution) error {
	f.createCalls++
	f.createContribution = contribution
	return f.createErr
}

func (f *fakeContributionService) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error) {
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

func (f *fakeContributionService) Delete(ctx context.Context, idStr string) error {
	f.deleteCalls++
	f.deleteID = idStr
	return f.deleteErr
}

func (f *fakeContributionService) Update(ctx context.Context, status string, idStr string) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = idStr
	return f.updateErr
}
func TestHandlePostContribution(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		createErr          error
		statusWanted       int
		errMsg             string
		wantCalls          int
		contributionWanted *models.Contribution
	}{
		{
			name:               "201 when service accepts the contribution",
			body:               `{"course_name":"A","link_url":"https://example.com","note":""}`,
			createErr:          nil,
			statusWanted:       http.StatusCreated,
			wantCalls:          1,
			contributionWanted: &models.Contribution{CourseName: "A", LinkURL: "https://example.com", Note: ""},
		},
		{
			name:               "201 accepts link_type from frontend",
			body:               `{"course_name":"A","link_url":"https://example.com","link_type":"drive","note":"[Type:drive] "}`,
			statusWanted:       http.StatusCreated,
			wantCalls:          1,
			contributionWanted: &models.Contribution{CourseName: "A", LinkURL: "https://example.com", LinkType: "drive", Note: "[Type:drive] "},
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
			body:         `{"course_name":"x","link_url":"https://x","note":""}`,
			createErr:    errs.ErrCourseNameAndLinkUrlRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course name and link URL are required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"course_name":"A","link_url":"https://example.com","note":""}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeContributionService := &fakeContributionService{createErr: tt.createErr}
			h := testHandler(t, withContribution(fakeContributionService))
			req := httptest.NewRequest(http.MethodPost, "/api/contributions", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handlePostContribution(rr, req)

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
				if fakeContributionService.createCalls != tt.wantCalls {
					t.Fatalf("service.Create calls: got %d want %d", fakeContributionService.createCalls, tt.wantCalls)
				}
				return
			}

			// from now on , only success cases pass
			if fakeContributionService.createCalls != tt.wantCalls {
				t.Fatalf("service.Create calls: got %d want %d", fakeContributionService.createCalls, tt.wantCalls)
			}
			if tt.contributionWanted == nil {
				t.Fatal("success case must set contributionWanted")
			}
			if !reflect.DeepEqual(fakeContributionService.createContribution, *tt.contributionWanted) {
				t.Fatalf("service.Create contribution: got %+v want %+v", fakeContributionService.createContribution, *tt.contributionWanted)
			}
			if rr.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", rr.Body.String())
			}
		})
	}
}

func TestHandleAdminGetContributions(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		listErr      error
		listResult   []models.Contribution
		statusWanted int
		errMsg       string
		wantLimit    int
		wantOffset   int
		wantQ        string
		wantStatus   string
	}{
		{
			name:         "accept empty status",
			target:       "/api/admin/contributions?limit=10&offset=10",
			listErr:      nil,
			wantStatus:   "",
			wantQ:        "",
			wantLimit:    10,
			wantOffset:   10,
			statusWanted: http.StatusOK,
		},
		{
			name:         "reject invalid status",
			target:       "/api/admin/contributions?limit=10&offset=10&status=open",
			listErr:      errs.ErrContributionInvalidStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be pending or approved",
		},
		{
			name:         "reject list invalid params",
			target:       "/api/admin/contributions?limit=10&offset=10&status=pending",
			listErr:      errs.ErrInvalidParams,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Limit should be between 1-100 and Offset >= 0",
		},
		{
			name:         "500 when service fails",
			target:       "/api/admin/contributions?limit=10&offset=10&status=pending",
			listErr:      errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
		},
		{
			name:         "accept pending status",
			target:       "/api/admin/contributions?limit=10&offset=10&status=pending",
			listResult:   []models.Contribution{{ID: 1, CourseName: "Linux", Status: "pending"}},
			statusWanted: http.StatusOK,
			wantLimit:    10,
			wantOffset:   10,
			wantStatus:   "pending",
		},
		{
			name:         "accept approved status with search",
			target:       "/api/admin/contributions?limit=50&offset=100&q=Linux&status=approved",
			listResult:   []models.Contribution{{ID: 2, CourseName: "Go", Status: "approved"}},
			statusWanted: http.StatusOK,
			wantLimit:    50,
			wantOffset:   100,
			wantQ:        "Linux",
			wantStatus:   "approved",
		},
		{
			name:         "default pagination and trimmed status",
			target:       "/api/admin/contributions?status=+pending+",
			listResult:   []models.Contribution{},
			statusWanted: http.StatusOK,
			wantLimit:    25,
			wantOffset:   0,
			wantStatus:   "pending",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeContributionService := &fakeContributionService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withContribution(fakeContributionService))
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetContributions(rr, req)

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
				if fakeContributionService.listCalls != 1 {
					t.Fatalf("service.List calls: got %d want 1", fakeContributionService.listCalls)
				}
				return
			}

			if fakeContributionService.listCalls != 1 {
				t.Fatalf("service.List calls: got %d want 1", fakeContributionService.listCalls)
			}
			if fakeContributionService.listLimit != tt.wantLimit || fakeContributionService.listOffset != tt.wantOffset || fakeContributionService.listQ != tt.wantQ || fakeContributionService.listStatus != tt.wantStatus {
				t.Fatalf("service.List args: got limit=%d offset=%d q=%q status=%q want limit=%d offset=%d q=%q status=%q",
					fakeContributionService.listLimit, fakeContributionService.listOffset, fakeContributionService.listQ, fakeContributionService.listStatus,
					tt.wantLimit, tt.wantOffset, tt.wantQ, tt.wantStatus)
			}
			var got []models.Contribution
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}

func TestHandleAdminUpdateContribution(t *testing.T) {
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
			name:         "400 invalid contribution id",
			pathID:       "abc",
			body:         `{"status":"pending"}`,
			updateErr:    errs.ErrContributionInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid contribution id",
			wantCalls:    1,
		},
		{
			name:         "404 contribution not found",
			pathID:       "10",
			body:         `{"status":"pending"}`,
			updateErr:    errs.ErrContributionNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Contribution not found",
			wantCalls:    1,
		},
		{
			name:         "400 status is required",
			pathID:       "10",
			body:         `{"status":""}`,
			updateErr:    errs.ErrStatusRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status is required",
			wantCalls:    1,
		},
		{
			name:         "400 invalid status",
			pathID:       "10",
			body:         `{"status":"open"}`,
			updateErr:    errs.ErrContributionInvalidStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be pending or approved",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"status":"pending"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept valid pending status",
			pathID:       "10",
			body:         `{"status":"pending"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "pending",
		},
		{
			name:         "accept valid approved status",
			pathID:       "10",
			body:         `{"status":"approved"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "approved",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeContributionService := &fakeContributionService{updateErr: tt.updateErr}
			h := testHandler(t, withContribution(fakeContributionService))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/contributions/"+tt.pathID, bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminUpdateContribution(rr, req)

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
				if fakeContributionService.updateCalls != tt.wantCalls {
					t.Fatalf("service.Update calls: got %d want %d", fakeContributionService.updateCalls, tt.wantCalls)
				}
				return
			}

			if fakeContributionService.updateCalls != tt.wantCalls {
				t.Fatalf("service.Update calls: got %d want %d", fakeContributionService.updateCalls, tt.wantCalls)
			}
			if fakeContributionService.updateID != tt.wantID || fakeContributionService.updateStatus != tt.wantStatus {
				t.Fatalf("service.Update args: got id=%q status=%q want id=%q status=%q",
					fakeContributionService.updateID, fakeContributionService.updateStatus, tt.wantID, tt.wantStatus)
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

func TestHandleAdminDeleteContribution(t *testing.T) {
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
			name:         "400 invalid contribution id",
			pathID:       "abc",
			deleteErr:    errs.ErrContributionInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid contribution id",
			wantCalls:    1,
		},
		{
			name:         "404 contribution not found",
			pathID:       "10",
			deleteErr:    errs.ErrContributionNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Contribution not found",
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
			fakeContributionService := &fakeContributionService{deleteErr: tt.deleteErr}
			h := testHandler(t, withContribution(fakeContributionService))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/contributions/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteContribution(rr, req)

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
				if fakeContributionService.deleteCalls != tt.wantCalls {
					t.Fatalf("service.Delete calls: got %d want %d", fakeContributionService.deleteCalls, tt.wantCalls)
				}
				return
			}

			if fakeContributionService.deleteCalls != tt.wantCalls {
				t.Fatalf("service.Delete calls: got %d want %d", fakeContributionService.deleteCalls, tt.wantCalls)
			}
			if fakeContributionService.deleteID != tt.wantID {
				t.Fatalf("service.Delete id: got %q want %q", fakeContributionService.deleteID, tt.wantID)
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
