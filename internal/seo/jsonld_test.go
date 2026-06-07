package seo

import (
	"strings"
	"testing"
)

func TestBuildCourseJSONLD(t *testing.T) {
	data := sampleCoursePageData()
	canonical := "https://example.com/course/nfa008"

	got := buildCourseJSONLD("https://example.com", data, canonical)
	if !strings.Contains(got, `application/ld+json`) {
		t.Fatalf("missing json-ld script tag: %q", got)
	}
	for _, want := range []string{
		"schema.org",
		"Course",
		"Bases de Données",
		"NFA008",
		"ItemList",
		"https://a.test",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("json-ld missing %q", want)
		}
	}
}
