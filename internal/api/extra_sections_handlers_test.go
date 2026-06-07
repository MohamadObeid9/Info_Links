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

type fakeExtraSectionService struct {
	listCalls   int
	listResult  []models.ExtraSection
	listErr     error
	createCalls int
	create      models.ExtraSection
	createErr   error
	updateCalls int
	update      models.ExtraSection
	updateID    string
	updateErr   error
	deleteCalls int
	deleteID    string
	deleteErr   error
}

func (f *fakeExtraSectionService) List(ctx context.Context) ([]models.ExtraSection, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeExtraSectionService) Create(ctx context.Context, section models.ExtraSection) error {
	f.createCalls++
	f.create = section
	return f.createErr
}

func (f *fakeExtraSectionService) Update(ctx context.Context, section models.ExtraSection, idStr string) error {
	f.updateCalls++
	f.update = section
	f.updateID = idStr
	return f.updateErr
}

func (f *fakeExtraSectionService) Delete(ctx context.Context, idStr string) error {
	f.deleteCalls++
	f.deleteID = idStr
	return f.deleteErr
}

func TestHandleAdminGetExtraSections(t *testing.T) {
	sample := []models.ExtraSection{
		{ID: 1, Title: "Python Resources", Icon: "🐍", DisplayOrder: 0},
		{ID: 2, Title: "Career", Icon: "💼", DisplayOrder: 1},
	}

	tests := []struct {
		name         string
		listErr      error
		listResult   []models.ExtraSection
		statusWanted int
		errMsg       string
		wantCalls    int
	}{
		{
			name:         "200 returns extra sections",
			listResult:   sample,
			statusWanted: http.StatusOK,
			wantCalls:    1,
		},
		{
			name:         "200 returns empty list",
			listResult:   []models.ExtraSection{},
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
			fake := &fakeExtraSectionService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withExtraSection(fake))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/extra_sections", nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetExtraSections(rr, req)

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

			var got []models.ExtraSection
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}

func TestHandleAdminPostExtraSection(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.ExtraSection
	}{
		{
			name:         "201 when service accepts the section",
			body:         `{"title":"Python Resources","icon":"🐍","display_order":0}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.ExtraSection{Title: "Python Resources", Icon: "🐍", DisplayOrder: 0},
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 when title is required",
			body:         `{"title":"","icon":"📁"}`,
			createErr:    errs.ErrExtraSectionTitleRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Extra section title is required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"title":"Python Resources","icon":"🐍"}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExtraSectionService{createErr: tt.createErr}
			h := testHandler(t, withExtraSection(fake))
			req := httptest.NewRequest(http.MethodPost, "/api/admin/extra_sections", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPostExtraSection(rr, req)

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
			if !reflect.DeepEqual(fake.create, *tt.resultWanted) {
				t.Fatalf("service.Create: got %+v want %+v", fake.create, *tt.resultWanted)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("status: got %q want ok", got["status"])
			}
		})
	}
}

func TestHandleAdminPatchExtraSection(t *testing.T) {
	section := models.ExtraSection{Title: "Updated", Icon: "📁", DisplayOrder: 1}

	tests := []struct {
		name          string
		pathID        string
		body          string
		updateErr     error
		statusWanted  int
		errMsg        string
		wantCalls     int
		wantID        string
		sectionWanted *models.ExtraSection
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
			name:         "400 invalid extra section id",
			pathID:       "abc",
			body:         `{"title":"Updated","icon":"📁"}`,
			updateErr:    errs.ErrExtraSectionInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid extra section id",
			wantCalls:    1,
		},
		{
			name:         "404 extra section not found",
			pathID:       "10",
			body:         `{"title":"Updated","icon":"📁"}`,
			updateErr:    errs.ErrExtraSectionNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Extra section not found",
			wantCalls:    1,
		},
		{
			name:         "400 when title is required",
			pathID:       "10",
			body:         `{"title":""}`,
			updateErr:    errs.ErrExtraSectionTitleRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Extra section title is required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"title":"Updated","icon":"📁"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:          "accept valid section",
			pathID:        "10",
			body:          `{"title":"Updated","icon":"📁","display_order":1}`,
			statusWanted:  http.StatusOK,
			wantCalls:     1,
			wantID:        "10",
			sectionWanted: &section,
		},
		{
			name:          "accept a valid id with spaces",
			pathID:        "  10  ",
			body:          `{"title":"Updated","icon":"📁","display_order":1}`,
			statusWanted:  http.StatusOK,
			wantCalls:     1,
			wantID:        "  10  ",
			sectionWanted: &section,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExtraSectionService{updateErr: tt.updateErr}
			h := testHandler(t, withExtraSection(fake))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/extra_sections/10", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPatchExtraSection(rr, req)

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
			if fake.updateID != tt.wantID {
				t.Fatalf("service.Update id: got %q want %q", fake.updateID, tt.wantID)
			}
			if !reflect.DeepEqual(fake.update, *tt.sectionWanted) {
				t.Fatalf("service.Update section: got %+v want %+v", fake.update, *tt.sectionWanted)
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

func TestHandleAdminDeleteExtraSection(t *testing.T) {
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
			name:         "400 invalid extra section id",
			pathID:       "abc",
			deleteErr:    errs.ErrExtraSectionInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid extra section id",
			wantCalls:    1,
		},
		{
			name:         "404 extra section not found",
			pathID:       "10",
			deleteErr:    errs.ErrExtraSectionNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Extra section not found",
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
			fake := &fakeExtraSectionService{deleteErr: tt.deleteErr}
			h := testHandler(t, withExtraSection(fake))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/extra_sections/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteExtraSection(rr, req)

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
