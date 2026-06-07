package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"infolinks-backend/internal/errs"
)

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
	Code       string
	Name       string
	Placements []CoursePlacement
	Links      []SEOLink
}

// ProgramPageData is a program and its courses for SEO.
type ProgramPageData struct {
	ID      int
	Name    string
	Slug    string
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

type postgresSEORepository struct {
	db *sql.DB
}

func NewPostgresSEORepository(db *sql.DB) SEORepository {
	return &postgresSEORepository{db: db}
}

func (r *postgresSEORepository) GetCoursePageByCode(ctx context.Context, code string) (*CoursePageData, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errs.ErrCourseNotFound
	}

	rows, err := r.db.QueryContext(ctx, getSEOCoursePlacementsQuery, code)
	if err != nil {
		return nil, fmt.Errorf("query seo course placements: %w", err)
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
			return nil, fmt.Errorf("scan seo course placement: %w", err)
		}
		placements = append(placements, p)
		courseIDs = append(courseIDs, p.CourseID)
		if displayName == "" {
			displayName = p.CourseName
			displayCode = p.Code
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo course placements rows: %w", err)
	}
	if len(placements) == 0 {
		return nil, errs.ErrCourseNotFound
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

func (r *postgresSEORepository) fetchLinksForCourses(ctx context.Context, courseIDs []int) ([]SEOLink, error) {
	if len(courseIDs) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(courseIDs))
	placeholders := make([]string, len(courseIDs))
	for i, id := range courseIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	q := fmt.Sprintf(listSEOLinksByCourseIDsQuery, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query seo links: %w", err)
	}
	defer rows.Close()

	var links []SEOLink
	seen := make(map[string]bool)
	for rows.Next() {
		var l SEOLink
		if err := rows.Scan(&l.ID, &l.Label, &l.URL, &l.Note, &l.ContentType, &l.LinkType); err != nil {
			return nil, fmt.Errorf("scan seo link: %w", err)
		}
		key := l.URL + "\x00" + l.Label
		if seen[key] {
			continue
		}
		seen[key] = true
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo links rows: %w", err)
	}
	return links, nil
}

func (r *postgresSEORepository) ListCourseCodesForSitemap(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, listSEOCourseCodesForSitemapQuery)
	if err != nil {
		return nil, fmt.Errorf("list seo course codes: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("scan seo course code: %w", err)
		}
		codes = append(codes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo course codes rows: %w", err)
	}
	return codes, nil
}

func (r *postgresSEORepository) ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]ProgramSitemapEntry, error) {
	rows, err := r.db.QueryContext(ctx, listSEOProgramsQuery)
	if err != nil {
		return nil, fmt.Errorf("list seo programs: %w", err)
	}
	defer rows.Close()

	var out []ProgramSitemapEntry
	for rows.Next() {
		var e ProgramSitemapEntry
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, fmt.Errorf("scan seo program: %w", err)
		}
		e.Slug = slugFn(e.Name)
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo programs rows: %w", err)
	}
	return out, nil
}

func (r *postgresSEORepository) ListCoursesIndex(ctx context.Context) ([]CourseIndexEntry, error) {
	rows, err := r.db.QueryContext(ctx, listSEOCoursesIndexQuery)
	if err != nil {
		return nil, fmt.Errorf("list seo courses index: %w", err)
	}
	defer rows.Close()

	var out []CourseIndexEntry
	for rows.Next() {
		var e CourseIndexEntry
		if err := rows.Scan(&e.Code, &e.Name, &e.ProgramName); err != nil {
			return nil, fmt.Errorf("scan seo courses index row: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo courses index rows: %w", err)
	}
	return out, nil
}

func (r *postgresSEORepository) GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*ProgramPageData, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, sql.ErrNoRows
	}

	rows, err := r.db.QueryContext(ctx, listSEOProgramsQuery)
	if err != nil {
		return nil, fmt.Errorf("list seo programs for slug: %w", err)
	}
	defer rows.Close()

	var prog *ProgramPageData
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan seo program slug row: %w", err)
		}
		if slugFn(name) == slug {
			prog = &ProgramPageData{ID: id, Name: name, Slug: slug}
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("seo program slug rows: %w", err)
	}
	if prog == nil {
		return nil, sql.ErrNoRows
	}

	crows, err := r.db.QueryContext(ctx, listSEOProgramCoursesQuery, prog.ID)
	if err != nil {
		return nil, fmt.Errorf("list seo program courses: %w", err)
	}
	defer crows.Close()

	for crows.Next() {
		var e ProgramCourseEntry
		if err := crows.Scan(&e.Code, &e.Name); err != nil {
			return nil, fmt.Errorf("scan seo program course: %w", err)
		}
		prog.Courses = append(prog.Courses, e)
	}
	if err := crows.Err(); err != nil {
		return nil, fmt.Errorf("seo program courses rows: %w", err)
	}
	return prog, nil
}
