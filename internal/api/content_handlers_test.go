package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"infolinks-backend/internal/errs"
)

type fakeContentService struct {
	getCalls        int
	invalidateCalls int
	getResult       []byte
	getErr          error
}

func (f *fakeContentService) Get(ctx context.Context) ([]byte, error) {
	f.getCalls++
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

func TestHandleGetContent_adminCacheHeader(t *testing.T) {
	sampleJSON := []byte(`{"programs":[]}`)
	fakeContent := &fakeContentService{getResult: sampleJSON}
	h := testHandler(t, withContent(fakeContent))

	adminToken := signTestToken(t, h.jwtSecret, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	nonAdminToken := signTestToken(t, h.jwtSecret, jwt.MapClaims{
		"admin": false,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	tests := []struct {
		name      string
		auth      string
		wantCache string
	}{
		{name: "admin token is not stored", auth: "Bearer " + adminToken, wantCache: contentCacheAdmin},
		{name: "non-admin token stays public", auth: "Bearer " + nonAdminToken, wantCache: contentCachePublic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
			req.Header.Set("Authorization", tt.auth)
			rr := httptest.NewRecorder()

			h.handleGetContent(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if cc := rr.Header().Get("Cache-Control"); cc != tt.wantCache {
				t.Fatalf("Cache-Control = %q, want %q", cc, tt.wantCache)
			}
			if !reflect.DeepEqual(rr.Body.Bytes(), sampleJSON) {
				t.Fatalf("body = %q, want %q", rr.Body.Bytes(), sampleJSON)
			}
		})
	}
	if fakeContent.getCalls != len(tests) {
		t.Fatalf("Get calls = %d, want %d", fakeContent.getCalls, len(tests))
	}
}
