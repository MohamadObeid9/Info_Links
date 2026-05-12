package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeReportService implements reportService for handler tests. Extend this
// struct with fields or funcs per method as you add tests for List/Update/Delete.
type fakeReportService struct {
	createErr error
}

func (f *fakeReportService) Create(ctx context.Context, report models.Report) error {
	_ = report
	return f.createErr
}

func (f *fakeReportService) List(ctx context.Context, limit, offset int, q, status string) ([]models.Report, error) {
	return nil, nil
}

func (f *fakeReportService) Delete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeReportService) Update(ctx context.Context, status, idStr string) error {
	return nil
}

func testReportsHandler(t *testing.T, svc *fakeReportService) *Handler {
	t.Helper()
	h, err := NewHandler(Dependencies{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:     []byte("test-jwt-secret"),
		ReportService: svc,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

// TestHandlePostReport shows the handler pattern: build request + httptest
// recorder, call the handler method, assert status and JSON body.
func TestHandlePostReport(t *testing.T) {
	t.Run("201 when service accepts the report", func(t *testing.T) {
		h := testReportsHandler(t, &fakeReportService{createErr: nil})

		body := `{"course_name":"A","link_url":"https://example.com","description":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.handlePostReport(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusCreated, rr.Body.String())
		}
	})

	t.Run("400 when service returns validation error", func(t *testing.T) {
		h := testReportsHandler(t, &fakeReportService{createErr: errs.ErrCourseNameAndLinkUrlAreRequired})

		body := `{"course_name":"x","link_url":"https://x","description":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.handlePostReport(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d body=%q", rr.Code, http.StatusBadRequest, rr.Body.String())
		}
		var got map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got["error"] == "" {
			t.Fatalf("expected error message in JSON: %v", got)
		}
	})
}
