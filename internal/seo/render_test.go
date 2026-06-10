package seo

import (
	"strings"
	"testing"

	"infolinks-backend/internal/repository"
)

func TestRenderCoursePage(t *testing.T) {
	html, err := renderCoursePage("https://example.com", sampleCoursePageData())
	if err != nil {
		t.Fatalf("renderCoursePage: %v", err)
	}
	for _, want := range []string{
		"Bases de Données",
		"NFA008",
		"schema.org",
		"https://example.com/course/nfa008",
		"Examens",
		"Ouvrir dans Info Links",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q", want)
		}
	}
}

func TestRenderProgramPage(t *testing.T) {
	data := &repository.ProgramPageData{
		ID:   1,
		Name: "Génie Info",
		Slug: "genie-info",
		Courses: []repository.ProgramCourseEntry{
			{Code: "nfa008", Name: "BDD"},
		},
	}
	html, err := renderProgramPage("https://example.com", data)
	if err != nil {
		t.Fatalf("renderProgramPage: %v", err)
	}
	for _, want := range []string{"Génie Info", "NFA008", "BDD", "/program/genie-info"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q", want)
		}
	}
}

func TestRenderCoursesIndex(t *testing.T) {
	entries := []repository.CourseIndexEntry{
		{Code: "nfa008", Name: "BDD", ProgramName: "Génie Info"},
	}
	html, err := renderCoursesIndex("https://example.com", entries)
	if err != nil {
		t.Fatalf("renderCoursesIndex: %v", err)
	}
	for _, want := range []string{"Tous les cours CNAM", "NFA008", "Génie Info", "/courses"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q", want)
		}
	}
}

func TestRender404(t *testing.T) {
	html, err := render404("https://example.com")
	if err != nil {
		t.Fatalf("render404: %v", err)
	}
	if !strings.Contains(html, "Cours introuvable") {
		t.Fatalf("html missing not found message: %q", html)
	}
}

func TestRender500(t *testing.T) {
	html, err := render500("https://example.com", "trace-abc-123")
	if err != nil {
		t.Fatalf("render500: %v", err)
	}
	if !strings.Contains(html, "Something went wrong") {
		t.Fatalf("html missing error heading: %q", html)
	}
	if !strings.Contains(html, "trace-abc-123") {
		t.Fatalf("html missing request id: %q", html)
	}
}

func TestExecuteTemplateUnknownName(t *testing.T) {
	_, err := executeTemplate("unknown", nil)
	if err == nil || !strings.Contains(err.Error(), "unknown template") {
		t.Fatalf("got err %v", err)
	}
}
