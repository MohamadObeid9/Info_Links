package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"infolinks-backend/internal/middleware"
)

func signTestToken(t *testing.T, secret []byte, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	return s
}

// fakeSupabaseToken builds a JWT that isAdmin() will accept (app_metadata.role=admin).
// The signature key doesn't matter since isAdmin uses ParseUnverified.
func fakeSupabaseToken(t *testing.T, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":          "test-user-uuid",
		"app_metadata": map[string]any{"role": role},
	})
	s, err := token.SignedString([]byte("fake-supabase-secret"))
	if err != nil {
		t.Fatalf("fakeSupabaseToken: %v", err)
	}
	return s
}

// supabaseServer spins up a test server simulating the Supabase password grant endpoint.
// It returns the given status and body for every request.
func supabaseServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestHandleLogin(t *testing.T) {
	adminToken := fakeSupabaseToken(t, "admin")
	userToken := fakeSupabaseToken(t, "user")

	tests := []struct {
		name            string
		body            string
		supabaseStatus  int
		supabaseBody    string
		statusWanted    int
		errMsg          string
		wantToken       bool
	}{
		{
			name:           "200 returns token for valid admin credentials",
			body:           `{"email":"admin@test.com","password":"secret"}`,
			supabaseStatus: http.StatusOK,
			supabaseBody:   `{"access_token":"` + adminToken + `"}`,
			statusWanted:   http.StatusOK,
			wantToken:      true,
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
		},
		{
			name:           "401 when supabase rejects credentials",
			body:           `{"email":"wrong@test.com","password":"wrong"}`,
			supabaseStatus: http.StatusBadRequest,
			supabaseBody:   `{"error":"invalid_grant","error_description":"Invalid login credentials"}`,
			statusWanted:   http.StatusUnauthorized,
			errMsg:         "Invalid credentials",
		},
		{
			name:           "403 when authenticated user is not admin",
			body:           `{"email":"user@test.com","password":"secret"}`,
			supabaseStatus: http.StatusOK,
			supabaseBody:   `{"access_token":"` + userToken + `"}`,
			statusWanted:   http.StatusForbidden,
			errMsg:         "Forbidden: Admin access required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)

			if tt.supabaseStatus != 0 {
				srv := supabaseServer(t, tt.supabaseStatus, tt.supabaseBody)
				h.supabaseURL = srv.URL
				h.httpClient = srv.Client()
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleLogin(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.errMsg != "" {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				return
			}

			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if !tt.wantToken {
				t.Fatal("success case must set wantToken")
			}
			if got["token"] == "" {
				t.Fatal("expected non-empty token")
			}

			parsed, err := jwt.Parse(got["token"], func(token *jwt.Token) (any, error) {
				return h.jwtSecret, nil
			})
			if err != nil || !parsed.Valid {
				t.Fatalf("token parse: err=%v valid=%v", err, parsed != nil && parsed.Valid)
			}
			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("expected map claims")
			}
			admin, ok := claims["admin"].(bool)
			if !ok || !admin {
				t.Fatalf("admin claim: got %v ok=%v", claims["admin"], ok)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	h := testHandler(t)
	secret := string(h.jwtSecret)

	validToken := signTestToken(t, h.jwtSecret, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	nonAdminToken := signTestToken(t, h.jwtSecret, jwt.MapClaims{
		"admin": false,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	wrongSecretToken := signTestToken(t, []byte("other-secret"), jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	expiredToken := signTestToken(t, h.jwtSecret, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(-time.Hour).Unix(),
	})

	tests := []struct {
		name         string
		authHeader   string
		statusWanted int
		errMsg       string
		wantNext     bool
	}{
		{
			name:         "401 when no token provided",
			authHeader:   "",
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Unauthorized: No token provided",
		},
		{
			name:         "401 when token is invalid",
			authHeader:   "Bearer not-a-jwt",
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Unauthorized: Invalid token",
		},
		{
			name:         "401 when token signed with wrong secret",
			authHeader:   "Bearer " + wrongSecretToken,
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Unauthorized: Invalid token",
		},
		{
			name:         "401 when token is expired",
			authHeader:   "Bearer " + expiredToken,
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Unauthorized: Invalid token",
		},
		{
			name:         "403 when admin claim is false",
			authHeader:   "Bearer " + nonAdminToken,
			statusWanted: http.StatusForbidden,
			errMsg:       "Forbidden: Admin access required",
		},
		{
			name:         "200 accept bearer admin token",
			authHeader:   "Bearer " + validToken,
			statusWanted: http.StatusOK,
			wantNext:     true,
		},
		{
			name:         "200 accept raw admin token without Bearer prefix",
			authHeader:   validToken,
			statusWanted: http.StatusOK,
			wantNext:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			wrapped := middleware.RequireAdmin(secret, func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/admin/courses", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			wrapped(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.errMsg != "" {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if nextCalled {
					t.Fatal("next handler should not be called")
				}
				return
			}

			if !tt.wantNext {
				t.Fatal("success case must set wantNext")
			}
			if !nextCalled {
				t.Fatal("next handler was not called")
			}
		})
	}
}
