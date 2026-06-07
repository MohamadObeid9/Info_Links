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

// fakeFeedbackService implements feedbackService for handler tests.
type fakeFeedbackService struct {
	createCalls    int
	createFeedback models.Feedback
	createErr      error

	listCalls  int
	listLimit  int
	listOffset int
	listQ      string
	listStatus string
	listResult []models.Feedback
	listErr    error

	deleteCalls int
	deleteID    string
	deleteErr   error

	updateCalls  int
	updateStatus string
	updateID     string
	updateErr    error
}

func (f *fakeFeedbackService) Create(ctx context.Context, feedback models.Feedback) error {
	f.createCalls++
	f.createFeedback = feedback
	return f.createErr
}

func (f *fakeFeedbackService) List(ctx context.Context, limit, offset int, q, status string) ([]models.Feedback, error) {
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

func (f *fakeFeedbackService) Delete(ctx context.Context, id string) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeFeedbackService) Update(ctx context.Context, status, idStr string) error {
	f.updateCalls++
	f.updateStatus = status
	f.updateID = idStr
	return f.updateErr
}

func TestHandlePostFeedback(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.Feedback
	}{
		{
			name:         "201 when service accepts the feedback",
			body:         `{"category":"performance","rating":5,"message":"nice"}`,
			createErr:    nil,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.Feedback{Category: "performance", Rating: 5, Message: "nice"},
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
			body:         `{"category":"","rating":0,"message":"nice"}`,
			createErr:    errs.ErrFeedbackCategoryAndRatingRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Category and rating are required",
			wantCalls:    1,
		},
		{
			name:         "400 when service returns category is malformed",
			body:         `{"category":"hello","rating":1,"message":"nice"}`,
			createErr:    errs.ErrFeedbackInvalidCategory,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Category must be one of the following : ui/ux or content or functionality or performance or accessibility",
			wantCalls:    1,
		},
		{
			name:         "400 when service returns rating is out of range",
			body:         `{"category":"performance","rating":100,"message":"nice"}`,
			createErr:    errs.ErrFeedbackInvalidRating,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Rating should be between 1 and 5",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"category":"performance","rating":5,"message":"nice"}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeFeedbackService := &fakeFeedbackService{createErr: tt.createErr}
			h := testHandler(t, withFeedback(fakeFeedbackService))
			req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handlePostFeedback(rr, req)

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
				if fakeFeedbackService.createCalls != tt.wantCalls {
					t.Fatalf("service.Create calls: got %d want %d", fakeFeedbackService.createCalls, tt.wantCalls)
				}
				return
			}

			// from now on , only success cases pass
			if fakeFeedbackService.createCalls != tt.wantCalls {
				t.Fatalf("service.Create calls: got %d want %d", fakeFeedbackService.createCalls, tt.wantCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(fakeFeedbackService.createFeedback, *tt.resultWanted) {
				t.Fatalf("service.Create feedback: got %+v want %+v", fakeFeedbackService.createFeedback, *tt.resultWanted)
			}
			if rr.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", rr.Body.String())
			}
		})
	}
}

func TestHandleAdminUpdateFeedback(t *testing.T) {
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
			name:         "400 invalid feedback id",
			pathID:       "abc",
			body:         `{"status":"read"}`,
			updateErr:    errs.ErrFeedbackInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid feedback id",
			wantCalls:    1,
		},
		{
			name:         "404 feedback not found",
			pathID:       "10",
			body:         `{"status":"read"}`,
			updateErr:    errs.ErrFeedbackNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Feedback not found",
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
			body:         `{"status":"pending"}`,
			updateErr:    errs.ErrFeedbackInvalidStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be new or read",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"status":"read"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept valid read status",
			pathID:       "10",
			body:         `{"status":"read"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "read",
		},
		{
			name:         "accept valid new status",
			pathID:       "10",
			body:         `{"status":"new"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			wantStatus:   "new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeFeedbackService := &fakeFeedbackService{updateErr: tt.updateErr}
			h := testHandler(t, withFeedback(fakeFeedbackService))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/feedback/"+tt.pathID, bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPatchFeedback(rr, req)

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
				if fakeFeedbackService.updateCalls != tt.wantCalls {
					t.Fatalf("service.Update calls: got %d want %d", fakeFeedbackService.updateCalls, tt.wantCalls)
				}
				return
			}

			if fakeFeedbackService.updateCalls != tt.wantCalls {
				t.Fatalf("service.Update calls: got %d want %d", fakeFeedbackService.updateCalls, tt.wantCalls)
			}
			if fakeFeedbackService.updateID != tt.wantID || fakeFeedbackService.updateStatus != tt.wantStatus {
				t.Fatalf("service.Update args: got id=%q status=%q want id=%q status=%q",
					fakeFeedbackService.updateID, fakeFeedbackService.updateStatus, tt.wantID, tt.wantStatus)
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

func TestHandleAdminDeleteFeedback(t *testing.T) {
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
			name:         "400 invalid feedback id",
			pathID:       "abc",
			deleteErr:    errs.ErrFeedbackInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid feedback id",
			wantCalls:    1,
		},
		{
			name:         "404 feedback not found",
			pathID:       "10",
			deleteErr:    errs.ErrFeedbackNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Feedback not found",
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
			fakeFeedbackService := &fakeFeedbackService{deleteErr: tt.deleteErr}
			h := testHandler(t, withFeedback(fakeFeedbackService))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/feedback/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteFeedback(rr, req)

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
				if fakeFeedbackService.deleteCalls != tt.wantCalls {
					t.Fatalf("service.Delete calls: got %d want %d", fakeFeedbackService.deleteCalls, tt.wantCalls)
				}
				return
			}

			if fakeFeedbackService.deleteCalls != tt.wantCalls {
				t.Fatalf("service.Delete calls: got %d want %d", fakeFeedbackService.deleteCalls, tt.wantCalls)
			}
			if fakeFeedbackService.deleteID != tt.wantID {
				t.Fatalf("service.Delete id: got %q want %q", fakeFeedbackService.deleteID, tt.wantID)
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

func TestHandleAdminGetFeedbacks(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		listErr      error
		listResult   []models.Feedback
		statusWanted int
		errMsg       string
		wantLimit    int
		wantOffset   int
		wantQ        string
		wantStatus   string
	}{
		{
			name:         "accept empty status",
			target:       "/api/admin/feedback?limit=10&offset=10",
			listErr:      nil,
			wantLimit:    10,
			wantOffset:   10,
			statusWanted: http.StatusOK,
		},
		{
			name:         "reject invalid status",
			target:       "/api/admin/feedback?limit=10&offset=10&status=pending",
			listErr:      errs.ErrFeedbackInvalidStatus,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Status must be new or read",
		},
		{
			name:         "reject list invalid params",
			target:       "/api/admin/feedback?limit=-10&offset=10&status=read",
			listErr:      errs.ErrInvalidParams,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Limit should be between 1-100 and Offset >= 0",
		},
		{
			name:         "500 when service fails",
			target:       "/api/admin/feedback?limit=10&offset=10&status=read",
			listErr:      errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
		},
		{
			name:         "accept read status",
			target:       "/api/admin/feedback?limit=10&offset=10&status=read",
			listResult:   []models.Feedback{{ID: 1, Category: "performance", Rating: 5, Status: "read"}},
			statusWanted: http.StatusOK,
			wantLimit:    10,
			wantOffset:   10,
			wantStatus:   "read",
		},
		{
			name:         "accept new status with search",
			target:       "/api/admin/feedback?limit=50&offset=100&q=Performance&status=new",
			listResult:   []models.Feedback{{ID: 2, Category: "performance", Rating: 5, Status: "new"}},
			statusWanted: http.StatusOK,
			wantLimit:    50,
			wantOffset:   100,
			wantQ:        "Performance",
			wantStatus:   "new",
		},
		{
			name:         "default pagination and trimmed status",
			target:       "/api/admin/feedback?status=+new+",
			listResult:   []models.Feedback{},
			statusWanted: http.StatusOK,
			wantLimit:    25,
			wantOffset:   0,
			wantStatus:   "new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeFeedbackService := &fakeFeedbackService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withFeedback(fakeFeedbackService))
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetFeedback(rr, req)

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
				if fakeFeedbackService.listCalls != 1 {
					t.Fatalf("service.List calls: got %d want 1", fakeFeedbackService.listCalls)
				}
				return
			}

			if fakeFeedbackService.listCalls != 1 {
				t.Fatalf("service.List calls: got %d want 1", fakeFeedbackService.listCalls)
			}
			if fakeFeedbackService.listLimit != tt.wantLimit || fakeFeedbackService.listOffset != tt.wantOffset || fakeFeedbackService.listQ != tt.wantQ || fakeFeedbackService.listStatus != tt.wantStatus {
				t.Fatalf("service.List args: got limit=%d offset=%d q=%q status=%q want limit=%d offset=%d q=%q status=%q",
					fakeFeedbackService.listLimit, fakeFeedbackService.listOffset, fakeFeedbackService.listQ, fakeFeedbackService.listStatus,
					tt.wantLimit, tt.wantOffset, tt.wantQ, tt.wantStatus)
			}
			var got []models.Feedback
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}
