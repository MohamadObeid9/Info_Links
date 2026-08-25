package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"infolinks-backend/internal/models"
)

type backup struct {
	models.ContentResponse
	ExportedAt    string          `json:"exported_at"`
	LinkClicks    json.RawMessage `json:"link_clicks"`
	skippedClicks int
}

func loadBackup(path string) (backup, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return backup{}, fmt.Errorf("read %s: %w", path, err)
	}

	var b backup
	if err := json.Unmarshal(raw, &b); err != nil {
		return backup{}, fmt.Errorf("parse %s: %w (need an admin backup with programs/years/courses arrays, not GET /api)", path, err)
	}
	if err := validateBackup(b); err != nil {
		return backup{}, fmt.Errorf("%s: %w", path, err)
	}
	if len(b.LinkClicks) > 0 && string(b.LinkClicks) != "null" {
		var clicks []json.RawMessage
		if err := json.Unmarshal(b.LinkClicks, &clicks); err == nil {
			b.skippedClicks = len(clicks)
		}
	}
	return b, nil
}

func validateBackup(b backup) error {
	if len(b.Programs) == 0 {
		return fmt.Errorf("no programs; export from Admin → Export (or GET /api/content), not GET /api")
	}
	for i, p := range b.Programs {
		if p.ID == 0 || p.Name == "" || p.Slug == "" {
			return fmt.Errorf("programs[%d] is not a course-tree row (need id, name, slug)", i)
		}
	}
	if len(b.Years) == 0 || len(b.Semesters) == 0 || len(b.Courses) == 0 {
		return fmt.Errorf("missing years, semesters, or courses")
	}
	for i, y := range b.Years {
		if y.ID == 0 || y.ProgramID == 0 || y.Name == "" {
			return fmt.Errorf("years[%d] is missing id, program_id, or name", i)
		}
	}
	for i, l := range b.Links {
		if l.ID == 0 || l.CourseID == nil || *l.CourseID == 0 || l.URL == "" {
			return fmt.Errorf("links[%d] is missing id, course_id, or url", i)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, b backup) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// CASCADE clears years/semesters/courses/links/extra_links/link_clicks/favorite_events.
	// users, reports, contributions, feedback, and page_views are left alone.
	if _, err := tx.ExecContext(ctx, `TRUNCATE TABLE programs, extra_sections RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("truncate content: %w", err)
	}

	if err := insertPrograms(ctx, tx, b.Programs); err != nil {
		return err
	}
	if err := insertYears(ctx, tx, b.Years); err != nil {
		return err
	}
	if err := insertSemesters(ctx, tx, b.Semesters); err != nil {
		return err
	}
	courses, links := canonicalizeCoursesAndLinks(b.Courses, b.Links)
	if err := insertCourses(ctx, tx, courses); err != nil {
		return err
	}
	if err := insertLinks(ctx, tx, links); err != nil {
		return err
	}
	if err := insertExtraSections(ctx, tx, b.ExtraSections); err != nil {
		return err
	}
	if err := insertExtraLinks(ctx, tx, b.ExtraLinks); err != nil {
		return err
	}

	tables := []string{"programs", "years", "semesters", "courses", "course_placements", "links", "extra_sections", "extra_links"}
	for _, table := range tables {
		q := fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), true)`,
			table, table,
		)
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("reset %s sequence: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func insertPrograms(ctx context.Context, tx *sql.Tx, rows []models.Program) error {
	const q = `INSERT INTO programs (id, name, slug, display_order) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.Name, r.Slug, r.DisplayOrder); err != nil {
			return fmt.Errorf("insert programs[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func insertYears(ctx context.Context, tx *sql.Tx, rows []models.Year) error {
	const q = `INSERT INTO years (id, program_id, name, display_order) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.ProgramID, r.Name, r.DisplayOrder); err != nil {
			return fmt.Errorf("insert years[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func insertSemesters(ctx context.Context, tx *sql.Tx, rows []models.Semester) error {
	const q = `INSERT INTO semesters (id, year_id, name, display_order) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.YearID, r.Name, r.DisplayOrder); err != nil {
			return fmt.Errorf("insert semesters[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func canonicalizeCoursesAndLinks(courses []models.Course, links []models.Link) ([]models.Course, []models.Link) {
	keepByCode := map[string]int{}
	idMap := map[int]int{}
	outCourses := make([]models.Course, 0, len(courses))
	for _, c := range courses {
		key := strings.ToLower(strings.TrimSpace(c.Code))
		if key != "" {
			if keep, ok := keepByCode[key]; ok {
				idMap[c.ID] = keep
				c.ID = keep
				outCourses = append(outCourses, c)
				continue
			}
			keepByCode[key] = c.ID
		}
		idMap[c.ID] = c.ID
		outCourses = append(outCourses, c)
	}

	seenURL := map[string]struct{}{}
	outLinks := make([]models.Link, 0, len(links))
	for _, l := range links {
		if l.CourseID == nil {
			continue
		}
		keep := idMap[*l.CourseID]
		if keep == 0 {
			keep = *l.CourseID
		}
		cid := keep
		l.CourseID = &cid
		ukey := fmt.Sprintf("%d|%s", cid, strings.ToLower(strings.TrimSpace(l.URL)))
		if _, ok := seenURL[ukey]; ok {
			continue
		}
		seenURL[ukey] = struct{}{}
		outLinks = append(outLinks, l)
	}
	return outCourses, outLinks
}

func insertCourses(ctx context.Context, tx *sql.Tx, rows []models.Course) error {
	const cq = `INSERT INTO courses (id, name, code, is_optional) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4)`
	const pq = `INSERT INTO course_placements (course_id, semester_id, display_order) VALUES ($1, $2, $3)`
	seen := map[int]struct{}{}
	seenPlacement := map[string]struct{}{}
	for i, r := range rows {
		if _, ok := seen[r.ID]; !ok {
			if _, err := tx.ExecContext(ctx, cq, r.ID, r.Name, r.Code, r.IsOptional); err != nil {
				return fmt.Errorf("insert courses[%d] id=%d: %w", i, r.ID, err)
			}
			seen[r.ID] = struct{}{}
		}
		if r.SemesterID <= 0 {
			continue
		}
		pkey := fmt.Sprintf("%d:%d", r.ID, r.SemesterID)
		if _, ok := seenPlacement[pkey]; ok {
			continue
		}
		seenPlacement[pkey] = struct{}{}
		if _, err := tx.ExecContext(ctx, pq, r.ID, r.SemesterID, r.DisplayOrder); err != nil {
			return fmt.Errorf("insert course_placements[%d] course_id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func insertLinks(ctx context.Context, tx *sql.Tx, rows []models.Link) error {
	const q = `INSERT INTO links (id, course_id, type, url, label, note, display_order, content_type) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.CourseID, r.Type, r.URL, r.Label, r.Note, r.DisplayOrder, r.ContentType); err != nil {
			return fmt.Errorf("insert links[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func insertExtraSections(ctx context.Context, tx *sql.Tx, rows []models.ExtraSection) error {
	const q = `INSERT INTO extra_sections (id, title, icon, display_order) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.Title, r.Icon, r.DisplayOrder); err != nil {
			return fmt.Errorf("insert extra_sections[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}

func insertExtraLinks(ctx context.Context, tx *sql.Tx, rows []models.ExtraLink) error {
	const q = `INSERT INTO extra_links (id, section_id, type, url, label, note, display_order, content_type) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	for i, r := range rows {
		if _, err := tx.ExecContext(ctx, q, r.ID, r.SectionID, r.Type, r.URL, r.Label, r.Note, r.DisplayOrder, r.ContentType); err != nil {
			return fmt.Errorf("insert extra_links[%d] id=%d: %w", i, r.ID, err)
		}
	}
	return nil
}
