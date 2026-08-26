package seo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"infolinks-backend/internal/repository"
)

func TestAcceptPrefersMarkdown(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{"", false},
		{"text/html", false},
		{"text/markdown", true},
		{"text/markdown; charset=utf-8", true},
		{"text/html, text/markdown;q=0.9", false},
		{"text/markdown, text/html", true},
		{"text/html;q=0.8, text/markdown;q=0.9", true},
		{"text/markdown;q=0", false},
		{"application/json", false},
	}
	for _, tt := range tests {
		if got := acceptPrefersMarkdown(tt.accept); got != tt.want {
			t.Errorf("acceptPrefersMarkdown(%q) = %v, want %v", tt.accept, got, tt.want)
		}
	}
}

func TestHandleCourseMarkdown(t *testing.T) {
	repo := &serviceFakeSEORepo{getCourseData: sampleCoursePageData()}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/course/nfa008", nil)
	req.SetPathValue("code", "nfa008")
	req.Header.Set("Accept", "text/markdown")
	rr := httptest.NewRecorder()

	h.HandleCourse(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/markdown") {
		t.Fatalf("Content-Type: got %q", ct)
	}
	if rr.Header().Get("Vary") != "Accept" {
		t.Fatalf("Vary: got %q", rr.Header().Get("Vary"))
	}
	if rr.Header().Get("x-markdown-tokens") == "" {
		t.Fatal("missing x-markdown-tokens")
	}
	body := rr.Body.String()
	if !strings.HasPrefix(body, "---\n") {
		t.Fatalf("expected YAML frontmatter, got: %s", body[:minLen(80, len(body))])
	}
	if !strings.Contains(body, "# ") {
		t.Fatal("expected markdown heading")
	}
	if !strings.Contains(body, "```json") {
		t.Fatal("expected JSON-LD fenced block")
	}
	if strings.Contains(body, "<html") || strings.Contains(body, "<script") {
		t.Fatal("markdown response should not include HTML chrome")
	}
}

func TestHandleCourseStillHTMLByDefault(t *testing.T) {
	repo := &serviceFakeSEORepo{getCourseData: sampleCoursePageData()}
	h := testSEOHandlerWithRepo(t, repo)
	req := httptest.NewRequest(http.MethodGet, "/course/nfa008", nil)
	req.SetPathValue("code", "nfa008")
	rr := httptest.NewRecorder()

	h.HandleCourse(rr, req)

	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("Vary") != "Accept" {
		t.Fatalf("Vary: got %q want Accept", rr.Header().Get("Vary"))
	}
}

func TestRenderProgramMarkdown(t *testing.T) {
	md, err := renderProgramMarkdown("https://example.com", &repository.ProgramPageData{
		Name: "Génie Info",
		Slug: "genie-info",
		Courses: []repository.ProgramCourseEntry{
			{Code: "nfa008", Name: "BDD"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "[NFA008](https://example.com/course/nfa008)") {
		t.Fatalf("missing course link: %s", md)
	}
}

func TestServeSPAMarkdown(t *testing.T) {
	rr := httptest.NewRecorder()
	ServeSPAMarkdown(rr, "https://example.com", "/")
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("Content-Type: %q", rr.Header().Get("Content-Type"))
	}
	if !strings.Contains(rr.Body.String(), "# Info Links") {
		t.Fatalf("body: %s", rr.Body.String())
	}
}
