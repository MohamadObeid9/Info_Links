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

type fakeExtraLinkService struct {
	listCalls   int
	listResult  []models.ExtraLink
	listErr     error
	createCalls int
	createLink  models.ExtraLink
	createErr   error
	updateCalls int
	updateLink  models.ExtraLink
	updateID    string
	updateErr   error
	deleteCalls int
	deleteID    string
	deleteErr   error
}

func (f *fakeExtraLinkService) List(ctx context.Context) ([]models.ExtraLink, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeExtraLinkService) Create(ctx context.Context, link models.ExtraLink) error {
	f.createCalls++
	f.createLink = link
	return f.createErr
}

func (f *fakeExtraLinkService) Update(ctx context.Context, link models.ExtraLink, idStr string) error {
	f.updateCalls++
	f.updateLink = link
	f.updateID = idStr
	return f.updateErr
}

func (f *fakeExtraLinkService) Delete(ctx context.Context, idStr string) error {
	f.deleteCalls++
	f.deleteID = idStr
	return f.deleteErr
}

func TestHandleAdminGetExtraLinks(t *testing.T) {
	sectionID := 1
	sample := []models.ExtraLink{
		{ID: 1, SectionID: &sectionID, Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 0},
	}

	tests := []struct {
		name         string
		listErr      error
		listResult   []models.ExtraLink
		statusWanted int
		errMsg       string
		wantCalls    int
	}{
		{
			name:         "200 returns extra links",
			listResult:   sample,
			statusWanted: http.StatusOK,
			wantCalls:    1,
		},
		{
			name:         "200 returns empty list",
			listResult:   []models.ExtraLink{},
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
			fake := &fakeExtraLinkService{listErr: tt.listErr, listResult: tt.listResult}
			h := testHandler(t, withExtraLink(fake))
			req := httptest.NewRequest(http.MethodGet, "/api/admin/extra_links", nil)
			rr := httptest.NewRecorder()

			h.handleAdminGetExtraLinks(rr, req)

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

			var got []models.ExtraLink
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.listResult) {
				t.Fatalf("response: got %+v want %+v", got, tt.listResult)
			}
		})
	}
}

func TestHandleAdminPostExtraLink(t *testing.T) {
	sectionID := 1
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.ExtraLink
	}{
		{
			name:         "201 when service accepts the link",
			body:         `{"section_id":1,"type":"telegram","url":"https://fake.test","label":"link 1","note":"","display_order":0}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.ExtraLink{SectionID: &sectionID, Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 0},
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 when url and label are required",
			body:         `{"section_id":1,"type":"telegram","url":"","label":"link 1"}`,
			createErr:    errs.ErrExtraLinkURLAndLabelRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Extra link url and label are required",
			wantCalls:    1,
		},
		{
			name:         "400 when section id is invalid",
			body:         `{"section_id":0,"type":"telegram","url":"https://fake.test","label":"link 1"}`,
			createErr:    errs.ErrExtraLinkInvalidSectionID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Extra link invalid section id",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"section_id":1,"type":"telegram","url":"https://fake.test","label":"link 1"}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExtraLinkService{createErr: tt.createErr}
			h := testHandler(t, withExtraLink(fake))
			req := httptest.NewRequest(http.MethodPost, "/api/admin/extra_links", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPostExtraLink(rr, req)

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
			if !reflect.DeepEqual(fake.createLink, *tt.resultWanted) {
				t.Fatalf("service.Create: got %+v want %+v", fake.createLink, *tt.resultWanted)
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

func TestHandleAdminPatchExtraLink(t *testing.T) {
	link := models.ExtraLink{Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil}

	tests := []struct {
		name         string
		pathID       string
		body         string
		updateErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
		linkWanted   *models.ExtraLink
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
			name:         "400 invalid extra link id",
			pathID:       "abc",
			body:         `{"label":"updated"}`,
			updateErr:    errs.ErrExtraLinkInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid extra link id",
			wantCalls:    1,
		},
		{
			name:         "404 extra link not found",
			pathID:       "10",
			body:         `{"label":"updated"}`,
			updateErr:    errs.ErrExtraLinkNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Extra link not found",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"label":"updated"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept valid link",
			pathID:       "10",
			body:         `{"type":"telegram","url":"https://fake.test","label":"link 1","note":""}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			linkWanted:   &link,
		},
		{
			name:         "accept a valid id with spaces",
			pathID:       "  10  ",
			body:         `{"type":"telegram","url":"https://fake.test","label":"link 1","note":""}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "  10  ",
			linkWanted:   &link,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExtraLinkService{updateErr: tt.updateErr}
			h := testHandler(t, withExtraLink(fake))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/extra_links/10", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPatchExtraLink(rr, req)

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
			if !reflect.DeepEqual(fake.updateLink, *tt.linkWanted) {
				t.Fatalf("service.Update link: got %+v want %+v", fake.updateLink, *tt.linkWanted)
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

func TestHandleAdminDeleteExtraLink(t *testing.T) {
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
			name:         "400 invalid extra link id",
			pathID:       "abc",
			deleteErr:    errs.ErrExtraLinkInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid extra link id",
			wantCalls:    1,
		},
		{
			name:         "404 extra link not found",
			pathID:       "10",
			deleteErr:    errs.ErrExtraLinkNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Extra link not found",
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
			fake := &fakeExtraLinkService{deleteErr: tt.deleteErr}
			h := testHandler(t, withExtraLink(fake))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/extra_links/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteExtraLink(rr, req)

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
