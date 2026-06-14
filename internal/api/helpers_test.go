package api

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

type mockPinger struct {
	err error
}

func (m mockPinger) Ping(ctx context.Context) error {
	return m.err
}

// handlerTestDeps holds fakes for NewHandler in tests. Nil fields use harmless defaults.
type handlerTestDeps struct {
	db           *mockPinger
	link         *fakeLinkService
	course       *fakeCourseService
	report       *fakeReportService
	content      *fakeContentService
	feedback     *fakeFeedbackService
	pageView     *fakePageViewService
	linkClick    *fakeLinkClickService
	contribution *fakeContributionService
	extraSection *fakeExtraSectionService
	extraLink    *fakeExtraLinkService
}

type testHandlerOption func(*handlerTestDeps)

func withDB(p *mockPinger) testHandlerOption {
	return func(d *handlerTestDeps) { d.db = p }
}

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
func withlink(s *fakeLinkService) testHandlerOption {
	return func(d *handlerTestDeps) { d.link = s }
}
func withlinkClick(s *fakeLinkClickService) testHandlerOption {
	return func(d *handlerTestDeps) { d.linkClick = s }
}
func withCourse(s *fakeCourseService) testHandlerOption {
	return func(d *handlerTestDeps) { d.course = s }
}
func withPageView(s *fakePageViewService) testHandlerOption {
	return func(d *handlerTestDeps) { d.pageView = s }
}
func withExtraSection(s *fakeExtraSectionService) testHandlerOption {
	return func(d *handlerTestDeps) { d.extraSection = s }
}
func withExtraLink(s *fakeExtraLinkService) testHandlerOption {
	return func(d *handlerTestDeps) { d.extraLink = s }
}

// testHandler builds a Handler for HTTP tests. Pass only the fakes you care about;
// omitted services get empty defaults that satisfy NewHandler.
func testHandler(t *testing.T, opts ...testHandlerOption) *Handler {
	t.Helper()

	deps := handlerTestDeps{
		db:           &mockPinger{},
		link:         &fakeLinkService{},
		report:       &fakeReportService{},
		course:       &fakeCourseService{},
		content:      &fakeContentService{},
		feedback:     &fakeFeedbackService{},
		pageView:     &fakePageViewService{},
		linkClick:    &fakeLinkClickService{},
		contribution: &fakeContributionService{},
		extraSection: &fakeExtraSectionService{},
		extraLink:    &fakeExtraLinkService{},
	}
	for _, opt := range opts {
		opt(&deps)
	}

	h, err := NewHandler(Dependencies{
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:           []byte("test-jwt-secret"),
		SupabaseURL:         "https://random.supabase.co",
		SupabaseAnonKey:     "a-random-generated-key",
		DB:                  deps.db,
		LinkService:         deps.link,
		ReportService:       deps.report,
		CourseService:       deps.course,
		ContentService:      deps.content,
		FeedbackService:     deps.feedback,
		PageViewService:     deps.pageView,
		LinkClickService:    deps.linkClick,
		ContributionService: deps.contribution,
		ExtraSectionService: deps.extraSection,
		ExtraLinkService:    deps.extraLink,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}
