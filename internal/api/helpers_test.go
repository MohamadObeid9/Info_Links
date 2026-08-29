package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/webbotauth"
)

// testStudentID is the student id the auth middleware would put in the context.
const testStudentID = 42

type mockPinger struct {
	err error
}

func (m mockPinger) Ping(ctx context.Context) error {
	return m.err
}

// handlerTestDeps holds fakes for NewHandler in tests. Nil fields use harmless defaults.
type handlerTestDeps struct {
	db           *mockPinger
	user         *fakeUserService
	analytics    *fakeAnalyticsService
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
	service      *fakeServiceService
}

type testHandlerOption func(*handlerTestDeps)

func withDB(p *mockPinger) testHandlerOption {
	return func(d *handlerTestDeps) { d.db = p }
}

func withUser(s *fakeUserService) testHandlerOption {
	return func(d *handlerTestDeps) { d.user = s }
}

func withAnalytics(s *fakeAnalyticsService) testHandlerOption {
	return func(d *handlerTestDeps) { d.analytics = s }
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
func withService(s *fakeServiceService) testHandlerOption {
	return func(d *handlerTestDeps) { d.service = s }
}

// testHandler builds a Handler for HTTP tests. Pass only the fakes you care about;
// omitted services get empty defaults that satisfy NewHandler.
func testHandler(t *testing.T, opts ...testHandlerOption) *Handler {
	t.Helper()

	deps := handlerTestDeps{
		db:           &mockPinger{},
		user:         &fakeUserService{},
		analytics:    &fakeAnalyticsService{},
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
		service:      &fakeServiceService{},
	}
	for _, opt := range opts {
		opt(&deps)
	}

	webBot, err := webbotauth.NewDirectory("test-jwt-secret", "https://example.com")
	if err != nil {
		t.Fatalf("webbotauth.NewDirectory: %v", err)
	}

	h, err := NewHandler(Dependencies{
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:           []byte("test-jwt-secret"),
		SiteBaseURL:         "https://example.com",
		SupabaseURL:         "https://random.supabase.co",
		SupabaseAnonKey:     "a-random-generated-key",
		WebBotAuth:          webBot,
		DB:                  deps.db,
		UserService:         deps.user,
		AnalyticsService:    deps.analytics,
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
		ServiceService:      deps.service,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

// jsonRequest builds a JSON request with no authenticated student.
func jsonRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// studentRequest builds a JSON request already carrying a student identity, the
// way the auth middleware hands it to the handler.
func studentRequest(method, target, body string) *http.Request {
	req := jsonRequest(method, target, body)
	return req.WithContext(middleware.ContextWithUser(req.Context(), testStudentID, false))
}
