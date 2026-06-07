package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestHandleLogin(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", "admin@test.com")
	t.Setenv("ADMIN_PASSWORD", "secret")

	tests := []struct {
		name         string
		body         string
		statusWanted int
		errMsg       string
		wantToken    bool
	}{
		{
			name:         "200 returns token for valid credentials",
			body:         `{"email":"admin@test.com","password":"secret"}`,
			statusWanted: http.StatusOK,
			wantToken:    true,
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
		},
		{
			name:         "401 for invalid email",
			body:         `{"email":"wrong@test.com","password":"secret"}`,
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Invalid credentials",
		},
		{
			name:         "401 for invalid password",
			body:         `{"email":"admin@test.com","password":"wrong"}`,
			statusWanted: http.StatusUnauthorized,
			errMsg:       "Invalid credentials",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t)
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
			name:         "403 when admin claim is false",
			authHeader:   "Bearer " + nonAdminToken,
			statusWanted: http.StatusForbidden,
			errMsg:       "Forbidden: Admin access required",
		},
		{
			name:         "accept bearer admin token",
			authHeader:   "Bearer " + validToken,
			statusWanted: http.StatusOK,
			wantNext:     true,
		},
		{
			name:         "accept raw admin token",
			authHeader:   validToken,
			statusWanted: http.StatusOK,
			wantNext:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			wrapped := h.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
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
