package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"infolinks-backend/internal/database"
)

var ErrCourseNotFound = errors.New("course not found")

// CoursePlacement is one program/year/semester location for a course code.
type CoursePlacement struct {
	CourseID     int
	CourseName   string
	Code         string
	IsOptional   bool
	ProgramID    int
	ProgramName  string
	YearName     string
	SemesterName string
}

// SEOLink is a resource link for SEO pages.
type SEOLink struct {
	ID          int
	Label       string
	URL         string
	Note        string
	ContentType string
	LinkType    string
}

// CoursePageData aggregates all placements and links for a course code.
type CoursePageData struct {
	Code        string
	Name        string
	Placements  []CoursePlacement
	Links       []SEOLink
}

// ProgramPageData is a program and its courses for SEO.
type ProgramPageData struct {
	ID    int
	Name  string
	Slug  string
	Courses []ProgramCourseEntry
}

// ProgramCourseEntry is a course row on a program page.
type ProgramCourseEntry struct {
	Code string
	Name string
}

// CourseIndexEntry is one row on /courses.
type CourseIndexEntry struct {
	Code        string
	Name        string
	ProgramName string
}

// ProgramSitemapEntry is a program for sitemap generation.
type ProgramSitemapEntry struct {
	ID   int
	Name string
	Slug string
}

// SEORepository reads data for server-rendered SEO pages.
type SEORepository struct {
	db *sql.DB
}

// NewSEORepository returns a repository using the global DB pool.
func NewSEORepository() *SEORepository {
	return &SEORepository{db: database.DB}
}

// GetCoursePageByCode returns placements and links for all courses matching code.
func (r *SEORepository) GetCoursePageByCode(ctx context.Context, code string) (*CoursePageData, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrCourseNotFound
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.name, c.code, c.is_optional,
		       p.id, p.name, y.name, s.name
		FROM courses c
		JOIN semesters s ON c.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		JOIN programs p ON y.program_id = p.id
		WHERE LOWER(TRIM(c.code)) = LOWER(TRIM($1))
		ORDER BY p.display_order, y.display_order, s.display_order, c.display_order
	`, code)
	if err != nil {
		return nil, fmt.Errorf("query placements: %w", err)
	}
	defer rows.Close()

	var placements []CoursePlacement
	var courseIDs []int
	displayName := ""
	displayCode := ""

	for rows.Next() {
		var p CoursePlacement
		if err := rows.Scan(
			&p.CourseID, &p.CourseName, &p.Code, &p.IsOptional,
			&p.ProgramID, &p.ProgramName, &p.YearName, &p.SemesterName,
		); err != nil {
			return nil, fmt.Errorf("scan placement: %w", err)
		}
		placements = append(placements, p)
		courseIDs = append(courseIDs, p.CourseID)
		if displayName == "" {
			displayName = p.CourseName
			displayCode = p.Code
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(placements) == 0 {
		return nil, ErrCourseNotFound
	}

	links, err := r.fetchLinksForCourses(ctx, courseIDs)
	if err != nil {
		return nil, err
	}

	return &CoursePageData{
		Code:       displayCode,
		Name:       displayName,
		Placements: placements,
		Links:      links,
	}, nil
}

func (r *SEORepository) fetchLinksForCourses(ctx context.Context, courseIDs []int) ([]SEOLink, error) {
	if len(courseIDs) == 0 {
		return nil, nil
	}
	// Build IN clause placeholders
	args := make([]interface{}, len(courseIDs))
	placeholders := make([]string, len(courseIDs))
	for i, id := range courseIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	q := fmt.Sprintf(`
		SELECT l.id, l.label, l.url, COALESCE(l.note, ''),
		       COALESCE(l.content_type, ''), COALESCE(l.type, '')
		FROM links l
		WHERE l.course_id IN (%s)
		ORDER BY l.display_order ASC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	var links []SEOLink
	seen := make(map[string]bool)
	for rows.Next() {
		var l SEOLink
		if err := rows.Scan(&l.ID, &l.Label, &l.URL, &l.Note, &l.ContentType, &l.LinkType); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		key := l.URL + "\x00" + l.Label
		if seen[key] {
			continue
		}
		seen[key] = true
		links = append(links, l)
	}
	return links, rows.Err()
}

// ListCourseCodesForSitemap returns distinct lowercased course codes.
func (r *SEORepository) ListCourseCodesForSitemap(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT LOWER(TRIM(code))
		FROM courses
		WHERE code IS NOT NULL AND TRIM(code) <> ''
		ORDER BY 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// ListProgramsForSitemap returns programs with URL slugs.
func (r *SEORepository) ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]ProgramSitemapEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name FROM programs ORDER BY display_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProgramSitemapEntry
	for rows.Next() {
		var e ProgramSitemapEntry
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, err
		}
		e.Slug = slugFn(e.Name)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListCoursesIndex returns one row per distinct course code for /courses.
func (r *SEORepository) ListCoursesIndex(ctx context.Context) ([]CourseIndexEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ON (LOWER(TRIM(c.code)))
		       LOWER(TRIM(c.code)), c.name, p.name
		FROM courses c
		JOIN semesters s ON c.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		JOIN programs p ON y.program_id = p.id
		WHERE c.code IS NOT NULL AND TRIM(c.code) <> ''
		ORDER BY LOWER(TRIM(c.code)), p.display_order, c.display_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CourseIndexEntry
	for rows.Next() {
		var e CourseIndexEntry
		if err := rows.Scan(&e.Code, &e.Name, &e.ProgramName); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetProgramBySlug finds a program whose slug matches.
func (r *SEORepository) GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*ProgramPageData, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, sql.ErrNoRows
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM programs ORDER BY display_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prog *ProgramPageData
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		if slugFn(name) == slug {
			prog = &ProgramPageData{ID: id, Name: name, Slug: slug}
			break
		}
	}
	if prog == nil {
		return nil, sql.ErrNoRows
	}

	crows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ON (LOWER(TRIM(c.code))) LOWER(TRIM(c.code)), c.name
		FROM courses c
		JOIN semesters s ON c.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		WHERE y.program_id = $1 AND c.code IS NOT NULL AND TRIM(c.code) <> ''
		ORDER BY LOWER(TRIM(c.code)), c.display_order
	`, prog.ID)
	if err != nil {
		return nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var e ProgramCourseEntry
		if err := crows.Scan(&e.Code, &e.Name); err != nil {
			return nil, err
		}
		prog.Courses = append(prog.Courses, e)
	}
	return prog, crows.Err()
}
