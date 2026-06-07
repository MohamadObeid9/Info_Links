package seo

import (
	"strings"
	"testing"

	"infolinks-backend/internal/repository"
)

func sampleCoursePageData() *repository.CoursePageData {
	return &repository.CoursePageData{
		Code: "nfa008",
		Name: "Bases de Données",
		Placements: []repository.CoursePlacement{
			{ProgramName: "Génie Info"},
			{ProgramName: "Licence Info"},
		},
		Links: []repository.SEOLink{
			{Label: "TD 1", URL: "https://a.test", ContentType: "td"},
			{Label: "Partiel", URL: "https://b.test", ContentType: "exams"},
			{Label: "Cours", URL: "https://c.test", ContentType: "cours,videos"},
		},
	}
}

func TestBuildCourseTitle(t *testing.T) {
	data := sampleCoursePageData()

	title := BuildCourseTitle(data.Name, data.Code, data.Links)
	if !strings.Contains(title, "Bases de Données") {
		t.Fatalf("title missing course name: %q", title)
	}
	if !strings.Contains(title, "NFA008") {
		t.Fatalf("title missing code: %q", title)
	}

	longName := strings.Repeat("Very Long Course Name ", 8)
	fallbackTitle := BuildCourseTitle(longName, "nfa999", nil)
	if strings.Contains(fallbackTitle, "—") {
		t.Fatalf("expected fallback title without type segment: %q", fallbackTitle)
	}

	typedTitle := BuildCourseTitle("Algo", "nfa010", []repository.SEOLink{{ContentType: "td"}})
	if !strings.Contains(typedTitle, "TD") {
		t.Fatalf("typed title missing content labels: %q", typedTitle)
	}
}

func TestBuildCourseDescription(t *testing.T) {
	data := sampleCoursePageData()

	desc := BuildCourseDescription(data)
	if !strings.Contains(desc, "NFA008") {
		t.Fatalf("description missing code: %q", desc)
	}
	if !strings.Contains(desc, "CNAM Liban") {
		t.Fatalf("description missing institution: %q", desc)
	}
	if len(desc) > 160 {
		t.Fatalf("description too long: %d chars", len(desc))
	}

	longData := &repository.CoursePageData{
		Code: "nfa010",
		Name: strings.Repeat("Long course name ", 20),
		Placements: []repository.CoursePlacement{
			{ProgramName: "Program A"},
			{ProgramName: "Program B"},
		},
		Links: []repository.SEOLink{
			{ContentType: "td,tp,cours,sessions,videos,exams"},
		},
	}
	longDesc := BuildCourseDescription(longData)
	if len(longDesc) > 160 {
		t.Fatalf("expected truncated description, got len %d", len(longDesc))
	}
}

func TestGroupLinksBySection(t *testing.T) {
	links := []repository.SEOLink{
		{Label: "TD", ContentType: "td"},
		{Label: "Exam", ContentType: "exams"},
		{Label: "Other", ContentType: ""},
	}

	grouped := GroupLinksBySection(links)
	if len(grouped["td"]) != 1 {
		t.Fatalf("td section: got %d links", len(grouped["td"]))
	}
	if len(grouped["examens"]) != 1 {
		t.Fatalf("examens section: got %d links", len(grouped["examens"]))
	}
	if len(grouped[SectionOther]) != 1 {
		t.Fatalf("other section: got %d links", len(grouped[SectionOther]))
	}
}

func TestPresentContentLabels(t *testing.T) {
	links := []repository.SEOLink{
		{ContentType: "td, exams"},
		{ContentType: "videos"},
		{ContentType: "td"},
	}

	labels := presentContentLabels(links)
	if len(labels) != 3 {
		t.Fatalf("got labels %v, want 3 unique labels", labels)
	}
}

func TestUniqueProgramNames(t *testing.T) {
	placements := []repository.CoursePlacement{
		{ProgramName: "A"},
		{ProgramName: "B"},
		{ProgramName: "A"},
	}
	got := uniqueProgramNames(placements)
	if len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Fatalf("got %v", got)
	}
}
