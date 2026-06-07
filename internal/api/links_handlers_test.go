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

type fakeLinkService struct {
	createCalls int
	createLink  models.Link
	createErr   error

	deleteCalls int
	deleteID    string
	deleteErr   error

	updateCalls int
	updateLink  models.Link
	updateID    string
	updateErr   error
}

func (f *fakeLinkService) Create(ctx context.Context, link models.Link) error {
	f.createCalls++
	f.createLink = link
	return f.createErr
}

func (f *fakeLinkService) Delete(ctx context.Context, idStr string) error {
	f.deleteCalls++
	f.deleteID = idStr
	return f.deleteErr
}

func (f *fakeLinkService) Update(ctx context.Context, link models.Link, idStr string) error {
	f.updateCalls++
	f.updateLink = link
	f.updateID = idStr
	return f.updateErr
}

func TestHandleAdminPostLink(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.Link
	}{
		{
			name:         "201 when service accepts the link",
			body:         `{"type":"Telegram","url":"https://fake.test","label":"link 1","note":"","display_order":5}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 5},
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
			body:         `{"type":"Telegram","url":"","label":"link 1"}`,
			createErr:    errs.ErrLinkURLAndLabelRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Link url and link label are required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"type":"Telegram","url":"https://fake.test","label":"link 1"}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeLinkService{createErr: tt.createErr}
			h := testHandler(t, withlink(fake))
			req := httptest.NewRequest(http.MethodPost, "/api/admin/links", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPostLink(rr, req)

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
			if !reflect.DeepEqual(fake.createLink, *tt.resultWanted) {
				t.Fatalf("service.Create link: got %+v want %+v", fake.createLink, *tt.resultWanted)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("status: got %q want %q", got["status"], "ok")
			}
		})
	}
}

func TestHandleAdminDeleteLink(t *testing.T) {
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
			name:         "400 invalid link id",
			pathID:       "abc",
			deleteErr:    errs.ErrLinkInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid link id",
			wantCalls:    1,
		},
		{
			name:         "404 link not found",
			pathID:       "10",
			deleteErr:    errs.ErrLinkNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Link not found",
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
			fake := &fakeLinkService{deleteErr: tt.deleteErr}
			h := testHandler(t, withlink(fake))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/links/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteLink(rr, req)

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

func TestHandleAdminPatchLink(t *testing.T) {
	link := models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil}

	tests := []struct {
		name         string
		pathID       string
		body         string
		updateErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
		linkWanted   *models.Link
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
			name:         "400 invalid link id",
			pathID:       "abc",
			body:         `{"label":"updated"}`,
			updateErr:    errs.ErrLinkInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid link id",
			wantCalls:    1,
		},
		{
			name:         "404 link not found",
			pathID:       "10",
			body:         `{"label":"updated"}`,
			updateErr:    errs.ErrLinkNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Link not found",
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
			body:         `{"type":"Telegram","url":"https://fake.test","label":"link 1","note":""}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			linkWanted:   &link,
		},
		{
			name:         "accept a valid id with spaces",
			pathID:       "  10  ",
			body:         `{"type":"Telegram","url":"https://fake.test","label":"link 1","note":""}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "  10  ",
			linkWanted:   &link,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeLinkService{updateErr: tt.updateErr}
			h := testHandler(t, withlink(fake))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/links/10", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPatchLink(rr, req)

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
			if tt.linkWanted == nil {
				t.Fatal("success case must set linkWanted")
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
