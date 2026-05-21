package seo

import "testing"

func TestClassifyLinkSection(t *testing.T) {
	tests := []struct {
		ct   string
		want string
	}{
		{"exams", "examens"},
		{"td,cours", "td"},
		{"cours", "cours"},
		{"videos", "videos"},
		{"", SectionOther},
		{"unknown", SectionOther},
	}
	for _, tc := range tests {
		got := ClassifyLinkSection(tc.ct)
		if got != tc.want {
			t.Errorf("ClassifyLinkSection(%q) = %q, want %q", tc.ct, got, tc.want)
		}
	}
}
