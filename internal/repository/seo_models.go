package repository

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
