package api

import (
	"io"
	"log/slog"
	"testing"
)

// handlerTestDeps holds fakes for NewHandler in tests. Nil fields use harmless defaults.
type handlerTestDeps struct {
	report       *fakeReportService
	content      *fakeContentService
	feedback     *fakeFeedbackService
	contribution *fakeContributionService
}

type testHandlerOption func(*handlerTestDeps)

func withReport(s *fakeReportService) testHandlerOption {
	return func(d *handlerTestDeps) { d.report = s }
}

func withContent(s *fakeContentService) testHandlerOption {
	return func(d *handlerTestDeps) { d.content = s }
}

func withFeedback(s *fakeFeedbackService) testHandlerOption {
	return func(d *handlerTestDeps) { d.feedback = s }
}

func withContribution(s *fakeContributionService) testHandlerOption {
	return func(d *handlerTestDeps) { d.contribution = s }
}

// testHandler builds a Handler for HTTP tests. Pass only the fakes you care about;
// omitted services get empty defaults that satisfy NewHandler.
func testHandler(t *testing.T, opts ...testHandlerOption) *Handler {
	t.Helper()

	deps := handlerTestDeps{
		report:       &fakeReportService{},
		content:      &fakeContentService{},
		feedback:     &fakeFeedbackService{},
		contribution: &fakeContributionService{},
	}
	for _, opt := range opts {
		opt(&deps)
	}

	h, err := NewHandler(Dependencies{
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:           []byte("test-jwt-secret"),
		ReportService:       deps.report,
		ContentService:      deps.content,
		FeedbackService:     deps.feedback,
		ContributionService: deps.contribution,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}
