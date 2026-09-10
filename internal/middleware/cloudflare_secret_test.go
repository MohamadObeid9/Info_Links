package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireCloudflareSecret_rejectsMissingHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := RequireCloudflareSecret("origin-secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusForbidden)
	}
}

func TestRequireCloudflareSecret_rejectsWrongHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	handler := RequireCloudflareSecret("origin-secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	req.Header.Set(cloudflareSecretHeader, "wrong-secret")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusForbidden)
	}
}

func TestRequireCloudflareSecret_allowsMatchingHeader(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireCloudflareSecret("origin-secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	req.Header.Set(cloudflareSecretHeader, "origin-secret")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestRequireCloudflareSecret_bypassesWhenSecretUnset(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireCloudflareSecret("", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestRequireCloudflareSecret_allowsExemptPathsWithoutHeader(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		t.Run(path, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireCloudflareSecret("origin-secret", next)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
			}
			if !called {
				t.Fatal("expected next handler to be called")
			}
		})
	}
}
