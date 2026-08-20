package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeLinkClickService struct {
	createCalls int
	createLC    models.LinkClick
	createErr   error

	listCalls  int
	listResult []models.LinkClick
	listErr    error
}

func (f *fakeLinkClickService) Create(ctx context.Context, lc models.LinkClick) error {
	f.createCalls++
	f.createLC = lc
	return f.createErr
}

func (f *fakeLinkClickService) List(ctx context.Context) ([]models.LinkClick, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func TestHandlePostLinkClick(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		noStudent    bool
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.LinkClick
	}{
		{
			name:         "201 when service accepts the link click",
			body:         `{"link_id":42}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.LinkClick{LinkID: &[]int{42}[0], UserID: testStudentID},
		},
		{
			name:         "201 when service accepts the extra link click",
			body:         `{"extra_link_id":42}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.LinkClick{ExtraLinkID: &[]int{42}[0], UserID: testStudentID},
		},
		{
			name:         "201 takes the user id from the token, not the body",
			body:         `{"link_id":42,"user_id":999}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.LinkClick{LinkID: &[]int{42}[0], UserID: testStudentID},
		},
		{
			name:         "401 when the student token is missing",
			body:         `{"link_id":42}`,
			noStudent:    true,
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Unauthorized: No token provided",
			wantCalls:    0,
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 when link id and extra link id are set",
			body:         `{"link_id":42, "extra_link_id":42}`,
			statusWanted: http.StatusBadRequest,
			createErr:    errs.ErrLinkClickLinkIDAndExtraLinkIDSet,
			errMsg:       "Link id and extra link id cannot be set at the same time",
			wantCalls:    1,
		},
		{
			name:         "400 when link id and extra link id are not set",
			body:         `{}`,
			statusWanted: http.StatusBadRequest,
			createErr:    errs.ErrLinkClickLinkIDAndExtraLinkIDRequired,
			errMsg:       "Link id and extra link id are required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"link_id":42}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeLinkClickService{createErr: tt.createErr}
			h := testHandler(t, withlinkClick(fake))
			req := studentRequest(http.MethodPost, "/api/link_clicks", tt.body)
			if tt.noStudent {
				req = jsonRequest(http.MethodPost, "/api/link_clicks", tt.body)
			}
			rr := httptest.NewRecorder()

			h.handlePostLinkClick(rr, req)

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
			if !reflect.DeepEqual(fake.createLC, *tt.resultWanted) {
				t.Fatalf("service.Create link click: got %+v want %+v", fake.createLC, *tt.resultWanted)
			}
			if rr.Body.Len() != 0 {
				t.Fatalf("expected empty body, got %q", rr.Body.String())
			}
		})
	}
}

func TestHandleAdminGetLinkClicks(t *testing.T) {
	sample := []models.LinkClick{
		{ID: 1, LinkID: &[]int{42}[0], ExtraLinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
		{ID: 2, LinkID: &[]int{99}[0], ExtraLinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
		{ID: 3, ExtraLinkID: &[]int{42}[0], LinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
		{ID: 4, ExtraLinkID: &[]int{99}[0], LinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
	}

	tests := []struct {
		name         string
		listErr      error
		listResult   []models.LinkClick
		statusWanted int
		errMsg       string
		wantCalls    int
	}{
		{
			name:         "200 returns link clicks",
			listResult:   sample,
			statusWanted: http.StatusOK,
			wantCalls:    1,
		},
		{
			name:         "200 returns empty list",
			listResult:   []models.LinkClick{},
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
			fake := &fakeLinkClickService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withlinkClick(fake))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/link_clicks", nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetLinkClicks(rr, req)

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

			var got []models.LinkClick
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}
