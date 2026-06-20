package seo

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/service"
)

func testSEOHandler(t *testing.T) *Handler {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	dbClient, err := database.New(os.Getenv("DATABASE_URL"), logger)
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { _ = dbClient.Close() })

	seoService := service.NewSEOService(repository.NewPostgresSEORepository(dbClient.DB))
	return NewHandler(logger, seoService, "https://example.com")
}

func testSEOHandlerWithRepo(t *testing.T, repo repository.SEORepository) *Handler {
	t.Helper()
	return NewHandler(slog.Default(), service.NewSEOService(repo), "https://example.com")
}

func serveSEO(
	h http.Handler,
	rr *httptest.ResponseRecorder,
	req *http.Request,
	requestID string,
) {
	if requestID != "" {
		req.Header.Set(middleware.HeaderRequestID, requestID)
	}
	middleware.RequestID(h).ServeHTTP(rr, req)
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil, "  https://example.com/  ")
	if h.baseURL != "https://example.com" {
		t.Fatalf("baseURL: got %q", h.baseURL)
	}
}

func TestHandleRobots(t *testing.T) {
	h := NewHandler(slog.Default(), nil, "https://example.com")
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rr := httptest.NewRecorder()
	h.HandleRobots(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Sitemap: https://example.com/sitemap.xml") {
		t.Errorf("robots missing sitemap: %s", body)
	}
	if !strings.Contains(body, "Disallow: /admin") {
		t.Error("robots missing admin disallow")
	}
}

func TestHandleSitemap(t *testing.T) {
	repo := &serviceFakeSEORepo{
		listCodes:    []string{"nfa008"},
		listPrograms: []repository.ProgramSitemapEntry{{ID: 1, Name: "Génie Info", Slug: "genie-info"}},
	}
	h := testSEOHandlerWithRepo(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()
	h.HandleSitemap(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("invalid sitemap xml")
	}
	if !strings.Contains(body, "https://example.com/courses") {
		t.Error("sitemap missing /courses")
	}
	if !strings.Contains(body, "https://example.com/about") {
		t.Error("sitemap missing /about")
	}
	if !strings.Contains(body, "https://example.com/course/nfa008") {
		t.Error("sitemap missing course url")
	}
}

func TestHandleSitemapServiceError(t *testing.T) {
	repo := &serviceFakeSEORepo{listCodesErr: errs.ErrDatabaseDown}
	h := testSEOHandlerWithRepo(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()
	serveSEO(http.HandlerFunc(h.HandleSitemap), rr, req, "sitemap-trace-1")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ref: sitemap-trace-1") {
		t.Fatalf("body missing request ref: %q", rr.Body.String())
	}
}

func TestHandleCourseEmptyCode(t *testing.T) {
	h := testSEOHandlerWithRepo(t, &serviceFakeSEORepo{})
	req := httptest.NewRequest(http.MethodGet, "/course/", nil)
	req.SetPathValue("code", "  ")
	rr := httptest.NewRecorder()

	h.HandleCourse(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Cours introuvable") {
		t.Fatalf("body: %q", rr.Body.String())
	}
}

func TestHandleCourseNotFound(t *testing.T) {
	repo := &serviceFakeSEORepo{getCourseErr: errs.ErrCourseNotFound}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/course/ZZZZNOTACODE999", nil)
	req.SetPathValue("code", "ZZZZNOTACODE999")
	rr := httptest.NewRecorder()
	h.HandleCourse(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleCourseSuccess(t *testing.T) {
	repo := &serviceFakeSEORepo{getCourseData: sampleCoursePageData()}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/course/nfa008", nil)
	req.SetPathValue("code", "nfa008")
	rr := httptest.NewRecorder()

	h.HandleCourse(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String()[:minLen(200, rr.Body.Len())])
	}
	if !strings.Contains(rr.Body.String(), "schema.org") {
		t.Error("expected JSON-LD")
	}
}

func TestHandleCourseServiceError(t *testing.T) {
	repo := &serviceFakeSEORepo{getCourseErr: errs.ErrDatabaseDown}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/course/nfa008", nil)
	req.SetPathValue("code", "nfa008")
	rr := httptest.NewRecorder()

	serveSEO(http.HandlerFunc(h.HandleCourse), rr, req, "course-trace-1")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "course-trace-1") {
		t.Fatalf("body missing request id: %q", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Something went wrong") {
		t.Fatalf("body missing error page: %q", rr.Body.String())
	}
}

func TestHandleProgramSuccess(t *testing.T) {
	repo := &serviceFakeSEORepo{
		getProgramData: &repository.ProgramPageData{
			ID: 1, Name: "Génie Info", Slug: "genie-info",
			Courses: []repository.ProgramCourseEntry{{Code: "nfa008", Name: "BDD"}},
		},
	}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/program/genie-info", nil)
	req.SetPathValue("slug", "genie-info")
	rr := httptest.NewRecorder()

	h.HandleProgram(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String()[:minLen(200, rr.Body.Len())])
	}
	if !strings.Contains(rr.Body.String(), "Génie Info") {
		t.Fatal("expected program name in body")
	}
}

func TestHandleProgramNotFound(t *testing.T) {
	repo := &serviceFakeSEORepo{getProgramErr: sql.ErrNoRows}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/program/missing", nil)
	req.SetPathValue("slug", "missing")
	rr := httptest.NewRecorder()

	h.HandleProgram(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
}

func TestHandleProgramEmptySlug(t *testing.T) {
	h := testSEOHandlerWithRepo(t, &serviceFakeSEORepo{})
	req := httptest.NewRequest(http.MethodGet, "/program/", nil)
	req.SetPathValue("slug", "  ")
	rr := httptest.NewRecorder()

	h.HandleProgram(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
}

func TestHandleCoursesIndexSuccess(t *testing.T) {
	repo := &serviceFakeSEORepo{
		listIndex: []repository.CourseIndexEntry{{Code: "nfa008", Name: "BDD", ProgramName: "Génie Info"}},
	}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	rr := httptest.NewRecorder()

	h.HandleCoursesIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Tous les cours CNAM") {
		t.Fatal("expected courses index heading")
	}
}

func TestHandleCoursesIndexServiceError(t *testing.T) {
	repo := &serviceFakeSEORepo{listIndexErr: errs.ErrDatabaseDown}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	rr := httptest.NewRecorder()

	serveSEO(http.HandlerFunc(h.HandleCoursesIndex), rr, req, "courses-trace-1")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "courses-trace-1") {
		t.Fatalf("body missing request id: %q", rr.Body.String())
	}
}

func TestXmlEscape(t *testing.T) {
	got := xmlEscape(`a&b<"c>'`)
	want := "a&amp;b&lt;&quot;c&gt;&apos;"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestHandleCourseFoundIntegration(t *testing.T) {
	h := testSEOHandler(t)
	codes, err := h.service.ListCourseCodesForSitemap(context.Background())
	if err != nil || len(codes) == 0 {
		t.Skip("no courses in database")
	}
	code := codes[0]
	req := httptest.NewRequest(http.MethodGet, "/course/"+code, nil)
	req.SetPathValue("code", code)
	rr := httptest.NewRecorder()
	h.HandleCourse(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String()[:minLen(200, rr.Body.Len())])
	}
}

type serviceFakeSEORepo struct {
	getCourseData  *repository.CoursePageData
	getCourseErr   error
	listCodes      []string
	listCodesErr   error
	listPrograms   []repository.ProgramSitemapEntry
	listProgramsErr error
	listIndex      []repository.CourseIndexEntry
	listIndexErr   error
	getProgramData *repository.ProgramPageData
	getProgramErr  error
}

func (f *serviceFakeSEORepo) GetCoursePageByCode(ctx context.Context, code string) (*repository.CoursePageData, error) {
	if f.getCourseErr != nil {
		return nil, f.getCourseErr
	}
	return f.getCourseData, nil
}

func (f *serviceFakeSEORepo) ListCourseCodesForSitemap(ctx context.Context) ([]string, error) {
	if f.listCodesErr != nil {
		return nil, f.listCodesErr
	}
	return f.listCodes, nil
}

func (f *serviceFakeSEORepo) ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]repository.ProgramSitemapEntry, error) {
	if f.listProgramsErr != nil {
		return nil, f.listProgramsErr
	}
	return f.listPrograms, nil
}

func (f *serviceFakeSEORepo) ListCoursesIndex(ctx context.Context) ([]repository.CourseIndexEntry, error) {
	if f.listIndexErr != nil {
		return nil, f.listIndexErr
	}
	return f.listIndex, nil
}

func (f *serviceFakeSEORepo) GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*repository.ProgramPageData, error) {
	if f.getProgramErr != nil {
		return nil, f.getProgramErr
	}
	return f.getProgramData, nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
