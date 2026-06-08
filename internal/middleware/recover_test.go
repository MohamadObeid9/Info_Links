package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecover_returns500OnPanic(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)

	Recover(slog.New(slog.NewTextHandler(io.Discard, nil)), next).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q want application/json", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if got["error"] != "Internal server error" {
		t.Fatalf("error: got %q want %q", got["error"], "Internal server error")
	}
}

func TestRecover_passesThroughNormalRequests(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/content", nil)

	Recover(slog.New(slog.NewTextHandler(io.Discard, nil)), next).ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusNoContent)
	}
}

func TestRecover_logsPanicWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", nil)
	req.Header.Set(HeaderRequestID, "trace-abc")

	handler := RequestID(Recover(logger, next))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusInternalServerError)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic recovered") {
		t.Fatalf("log: expected panic recovered message, got %q", logOutput)
	}
	if !strings.Contains(logOutput, "trace-abc") {
		t.Fatalf("log: expected request_id trace-abc, got %q", logOutput)
	}
	if !strings.Contains(logOutput, "boom") {
		t.Fatalf("log: expected panic value, got %q", logOutput)
	}
	if !strings.Contains(logOutput, "/api/feedback") {
		t.Fatalf("log: expected path, got %q", logOutput)
	}
}

func TestRecover_nilLoggerUsesDefault(t *testing.T) {
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Recover(nil, next).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusInternalServerError)
	}
}
