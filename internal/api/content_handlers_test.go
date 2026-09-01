package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"infolinks-backend/internal/errs"
)

type fakeContentService struct {
	getCalls         int
	getUncachedCalls int
	invalidateCalls  int
	getResult        []byte
	getErr           error
}

func (f *fakeContentService) Get(ctx context.Context) ([]byte, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeContentService) GetUncached(ctx context.Context) ([]byte, error) {
	f.getUncachedCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeContentService) Invalidate() {
	f.invalidateCalls++
}

func TestHandleGetContent(t *testing.T) {
	sampleJSON := []byte(`{"programs":[],"years":[],"semesters":[],"courses":[],"links":[],"extra_sections":[],"extra_links":[]}`)

	tests := []struct {
		name         string
		getResult    []byte
		getErr       error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantBody     []byte
	}{
		{
			name:         "200 returns navigation json",
			getResult:    sampleJSON,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantBody:     sampleJSON,
		},
		{
			name:         "500 when service fails",
			getErr:       errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeContent := &fakeContentService{getResult: tt.getResult, getErr: tt.getErr}
			h := testHandler(t, withContent(fakeContent))
			req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
			rr := httptest.NewRecorder()

			h.handleGetContent(rr, req)

			if fakeContent.getCalls != tt.wantCalls {
				t.Fatalf("get calls = %d, want %d", fakeContent.getCalls, tt.wantCalls)
			}
			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg != "" {
				var body map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body["error"] != tt.errMsg {
					t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
				}
				if cc := rr.Header().Get("Cache-Control"); strings.Contains(cc, contentCachePublic) {
					t.Fatalf("error response must not be publicly cached, Cache-Control=%q", cc)
				}
				return
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
			if cc := rr.Header().Get("Cache-Control"); cc != contentCachePublic {
				t.Fatalf("Cache-Control = %q, want %q", cc, contentCachePublic)
			}
			if !reflect.DeepEqual(rr.Body.Bytes(), tt.wantBody) {
				t.Fatalf("body = %q, want %q", rr.Body.Bytes(), tt.wantBody)
			}
		})
	}
}

func TestHandleGetAdminContent(t *testing.T) {
	sampleJSON := []byte(`{"programs":[]}`)
	fakeContent := &fakeContentService{getResult: sampleJSON}
	h := testHandler(t, withContent(fakeContent))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/content", nil)
	rr := httptest.NewRecorder()

	h.handleGetAdminContent(rr, req)

	if fakeContent.getUncachedCalls != 1 {
		t.Fatalf("GetUncached calls = %d, want 1", fakeContent.getUncachedCalls)
	}
	if fakeContent.getCalls != 0 {
		t.Fatalf("Get calls = %d, want 0 (admin must bypass the student cache)", fakeContent.getCalls)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != contentCacheAdmin {
		t.Fatalf("Cache-Control = %q, want %q", cc, contentCacheAdmin)
	}
	if !reflect.DeepEqual(rr.Body.Bytes(), sampleJSON) {
		t.Fatalf("body = %q, want %q", rr.Body.Bytes(), sampleJSON)
	}
}
