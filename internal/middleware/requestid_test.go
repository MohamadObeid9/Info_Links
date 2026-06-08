package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_generatesWhenMissing(t *testing.T) {
	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)

	RequestID(next).ServeHTTP(rr, req)

	if gotID == "" {
		t.Fatal("expected request ID in context")
	}
	if rr.Header().Get(HeaderRequestID) != gotID {
		t.Fatalf("response header: got %q want %q", rr.Header().Get(HeaderRequestID), gotID)
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusNoContent)
	}
}

func TestRequestID_reusesClientHeader(t *testing.T) {
	const clientID = "client-trace-123"

	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = RequestIDFromContext(r.Context())
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, clientID)

	RequestID(next).ServeHTTP(rr, req)

	if gotID != clientID {
		t.Fatalf("context ID: got %q want %q", gotID, clientID)
	}
	if rr.Header().Get(HeaderRequestID) != clientID {
		t.Fatalf("response header: got %q want %q", rr.Header().Get(HeaderRequestID), clientID)
	}
}

func TestRequestID_rejectsInvalidHeader(t *testing.T) {
	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = RequestIDFromContext(r.Context())
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, "bad id!!!")

	RequestID(next).ServeHTTP(rr, req)

	if gotID == "" {
		t.Fatal("expected generated request ID in context")
	}
	if gotID == "bad id!!!" {
		t.Fatal("expected invalid client ID to be replaced")
	}
}

func TestRequestIDWithLogging_recordsStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	RequestIDWithLogging(slog.New(slog.NewTextHandler(io.Discard, nil)), next).ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusTeapot)
	}
	if rr.Header().Get(HeaderRequestID) == "" {
		t.Fatal("expected response request ID header")
	}
}

func TestIsValidRequestID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"abc-123", true},
		{"CLIENT-TRACE", true},
		{"", false},
		{"bad id", false},
		{"bad!!!", false},
		{string(make([]byte, 129)), false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := isValidRequestID(tt.id); got != tt.want {
				t.Fatalf("isValidRequestID(%q): got %v want %v", tt.id, got, tt.want)
			}
		})
	}
}
