package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"infolinks-backend/internal/models"
)

func TestValidateBackup(t *testing.T) {
	courseID := 1
	valid := backup{ContentResponse: models.ContentResponse{
		Programs:  []models.Program{{ID: 1, Name: "Licence Info", Slug: "licence"}},
		Years:     []models.Year{{ID: 1, ProgramID: 1, Name: "1ère Année"}},
		Semesters: []models.Semester{{ID: 1, YearID: 1, Name: "Semestre 1"}},
		Courses:   []models.Course{{ID: 1, SemesterID: 1, Name: "Java 1", Code: "NFA031"}},
		Links:     []models.Link{{ID: 1, CourseID: &courseID, Type: "drive", URL: "https://example.com"}},
	}}

	if err := validateBackup(valid); err != nil {
		t.Fatalf("valid backup: %v", err)
	}

	welcome := valid
	welcome.Programs = []models.Program{{ID: 0, Name: "", Slug: ""}}
	if err := validateBackup(welcome); err == nil {
		t.Fatal("welcome-payload programs: want error")
	}

	empty := backup{}
	if err := validateBackup(empty); err == nil {
		t.Fatal("empty backup: want error")
	}
}

func TestLoadBackup(t *testing.T) {
	courseID := 6
	payload := map[string]any{
		"exported_at": "2026-04-18T22:42:11.635Z",
		"programs":    []map[string]any{{"id": 1, "name": "Licence Info", "slug": "licence", "display_order": 1}},
		"years":       []map[string]any{{"id": 1, "program_id": 1, "name": "1ère Année", "display_order": 1}},
		"semesters":   []map[string]any{{"id": 1, "year_id": 1, "name": "Semestre 1", "display_order": 1}},
		"courses":     []map[string]any{{"id": 6, "semester_id": 1, "name": "Java 1", "code": "NFA031", "display_order": 1, "is_optional": false}},
		"links": []map[string]any{{
			"id": 1, "course_id": courseID, "type": "telegram",
			"url": "https://t.me/example", "label": "Link 1", "note": "", "display_order": 1,
		}},
		"extra_sections": []any{},
		"extra_links":    []any{},
		"link_clicks":    []map[string]any{{"id": 1, "link_id": nil}},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "backup.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := loadBackup(path)
	if err != nil {
		t.Fatalf("loadBackup: %v", err)
	}
	if got.Programs[0].Slug != "licence" {
		t.Fatalf("slug = %q, want licence", got.Programs[0].Slug)
	}
	if got.skippedClicks != 1 {
		t.Fatalf("skippedClicks = %d, want 1", got.skippedClicks)
	}
}

func TestGuardDSN(t *testing.T) {
	if err := guardDSN(defaultLocalDSN, false); err != nil {
		t.Fatalf("local dsn: %v", err)
	}
	remote := "postgres://postgres:postgres@db.example.supabase.co:5432/postgres"
	if err := guardDSN(remote, false); err == nil {
		t.Fatal("remote dsn without -allow-remote: want error")
	}
	if err := guardDSN(remote, true); err != nil {
		t.Fatalf("remote dsn with -allow-remote: %v", err)
	}
}
