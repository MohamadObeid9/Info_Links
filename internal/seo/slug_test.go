package seo

import "testing"

func TestProgramSlug(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Licence Génie Informatique", "licence-genie-informatique"},
		{"Master AISL", "master-aisl"},
		{"  IRSM  ", "irsm"},
		{"", ""},
	}
	for _, tc := range tests {
		got := ProgramSlug(tc.in)
		if got != tc.want {
			t.Errorf("ProgramSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
