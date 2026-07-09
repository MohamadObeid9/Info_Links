package models

// Program represents a major or field of study
type Program struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int    `json:"display_order"`
}

// Year represents an academic year within a program
type Year struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProgramID    int    `json:"program_id"`
	DisplayOrder int    `json:"display_order"`
}

// Semester represents a semester within an academic year
type Semester struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	YearID       int    `json:"year_id"`
	DisplayOrder int    `json:"display_order"`
}

// Course represents a class within a semester
type Course struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	IsOptional   bool   `json:"is_optional"`
	SemesterID   int    `json:"semester_id"`
	DisplayOrder int    `json:"display_order"`
}

// CoursePatch represents an updating course
type CoursePatch struct {
	Name       *string `json:"name"`
	Code       *string `json:"code"`
	IsOptional *bool   `json:"is_optional"`
	SemesterID *int    `json:"semester_id"`
}

// Link represents a useful resource for a course or extra section
type Link struct {
	ID           int     `json:"id"`
	Type         string  `json:"type"`
	Label        string  `json:"label"`
	URL          string  `json:"url"`
	Note         string  `json:"note"`
	ContentType  *string `json:"content_type"`
	DisplayOrder int     `json:"display_order"`
	CourseID     *int    `json:"course_id,omitempty"`
}

// ExtraSection represents a non-course category of links
type ExtraSection struct {
	ID           int    `json:"id"`
	Icon         string `json:"icon"` // For course links
	Title        string `json:"title"`
	DisplayOrder int    `json:"display_order"`
}

// Extra Link represents a useful resource for an additional course or extra section
type ExtraLink struct {
	ID           int     `json:"id"`
	URL          string  `json:"url"`
	Note         string  `json:"note"`
	Type         string  `json:"type"`
	Label        string  `json:"label"`
	ContentType  *string `json:"content_type"`
	DisplayOrder int     `json:"display_order"`
	SectionID    *int    `json:"section_id,omitempty"`
}

// Report represents a reported issue with a link
type Report struct {
	ID          int    `json:"id"`
	Status      string `json:"status"` // open, resolved
	LinkURL     string `json:"link_url"`
	CreatedAt   string `json:"created_at"`
	CourseName  string `json:"course_name"`
	Description string `json:"description"`
}

// Contribution represents a user-submitted link
type Contribution struct {
	ID         int    `json:"id"`
	Note       string `json:"note"`
	Status     string `json:"status"` // pending, approved
	LinkURL    string `json:"link_url"`
	LinkType   string `json:"link_type,omitempty"` // accepted on POST; stored in note as [Type:...]
	CreatedAt  string `json:"created_at"`
	CourseName string `json:"course_name"`
}

// Feedback represents user feedback
type Feedback struct {
	ID        int    `json:"id"`
	Rating    int    `json:"rating"`
	Status    string `json:"status"` // new, read
	Message   string `json:"message"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// PageView tracks site visits
type PageView struct {
	ID        int    `json:"id"`
	Page      string `json:"page"`
	VisitedAt string `json:"visited_at"`
}

// LinkClick tracks clicks on specific links
type LinkClick struct {
	ID          int    `json:"id"`
	LinkID      *int   `json:"link_id,omitempty"`
	ExtraLinkID *int   `json:"extra_link_id,omitempty"`
	ClickedAt   string `json:"clicked_at"`
}

// ContentResponse is the big JSON object we send to the frontend.
type ContentResponse struct {
	Links         []Link         `json:"links"`
	Years         []Year         `json:"years"`
	Courses       []Course       `json:"courses"`
	Programs      []Program      `json:"programs"`
	Semesters     []Semester     `json:"semesters"`
	ExtraLinks    []ExtraLink    `json:"extra_links"`
	ExtraSections []ExtraSection `json:"extra_sections"`
}
