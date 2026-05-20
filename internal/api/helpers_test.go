package api

import (
	"io"
	"log/slog"
	"testing"
)

func testHandler(t *testing.T, fakeReportService *fakeReportService, fakeContributionService *fakeContributionService) *Handler {
	t.Helper()
	h, err := NewHandler(Dependencies{
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:           []byte("test-jwt-secret"),
		ReportService:       fakeReportService,
		ContributionService: fakeContributionService,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}
