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

type fakePageViewService struct {
	createCalls int
	createPV    models.PageView
	createErr   error

	listCalls  int
	listResult []models.PageView
	listErr    error
}

func (f *fakePageViewService) Create(ctx context.Context, pv models.PageView) error {
	f.createCalls++
	f.createPV = pv
	return f.createErr
}

func (f *fakePageViewService) List(ctx context.Context) ([]models.PageView, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func TestHandlePostPageView(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.PageView
	}{
		{
			name:         "201 when service accepts the page view",
			body:         `{"page":"home"}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.PageView{Page: "home"},
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "500 when service fails",
			body:         `{"page":"home"}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakePageViewService{createErr: tt.createErr}
			h := testHandler(t, withPageView(fake))
			req := httptest.NewRequest(http.MethodPost, "/api/page_views", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handlePostPageView(rr, req)

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

			if fake.createCalls != tt.wantCalls {
				t.Fatalf("service.Create calls: got %d want %d", fake.createCalls, tt.wantCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(fake.createPV, *tt.resultWanted) {
				t.Fatalf("service.Create page view: got %+v want %+v", fake.createPV, *tt.resultWanted)
			}
			if rr.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", rr.Body.String())
			}
		})
	}
}

func TestHandleAdminGetPageViews(t *testing.T) {
	sample := []models.PageView{
		{ID: 1, Page: "home", VisitedAt: "2024-01-01T00:00:00Z"},
		{ID: 2, Page: "admin", VisitedAt: "2024-02-01T00:00:00Z"},
	}

	tests := []struct {
		name         string
		listErr      error
		listResult   []models.PageView
		statusWanted int
		errMsg       string
		wantCalls    int
	}{
		{
			name:         "200 returns page views",
			listResult:   sample,
			statusWanted: http.StatusOK,
			wantCalls:    1,
		},
		{
			name:         "200 returns empty list",
			listResult:   []models.PageView{},
			statusWanted: http.StatusOK,
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			listErr:      errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakePageViewService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withPageView(fake))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/page_views", nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetPageViews(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if fake.listCalls != tt.wantCalls {
				t.Fatalf("service.List calls: got %d want %d", fake.listCalls, tt.wantCalls)
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				return
			}

			var got []models.PageView
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}
