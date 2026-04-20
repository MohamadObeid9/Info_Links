package models

// Program represents a major or field of study
type Program struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
}

// Year represents an academic year within a program
type Year struct {
	ID           int    `json:"id"`
	ProgramID    int    `json:"program_id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
}

// Semester represents a semester within an academic year
type Semester struct {
	ID           int    `json:"id"`
	YearID       int    `json:"year_id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
}

// Course represents a class within a semester
type Course struct {
	ID           int    `json:"id"`
	SemesterID   int    `json:"semester_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	IsOptional   bool   `json:"is_optional"`
	DisplayOrder int    `json:"display_order"`
}

// Link represents a useful resource for a course or extra section
type Link struct {
	ID           int     `json:"id"`
	CourseID     *int    `json:"course_id,omitempty"` // For course links
	SectionID    *int    `json:"section_id,omitempty"` // For extra links
	Type         string  `json:"type"`
	Label        string  `json:"label"`
	URL          string  `json:"url"`
	Note         string  `json:"note"`
	ContentType  *string `json:"content_type"`
	DisplayOrder int     `json:"display_order"`
}

// ExtraSection represents a non-course category of links
type ExtraSection struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Icon         string `icon:"icon"`
	DisplayOrder int    `json:"display_order"`
}

// Report represents a reported issue with a link
type Report struct {
	ID          int    `json:"id"`
	CourseName  string `json:"course_name"`
	LinkURL     string `json:"link_url"`
	Description string `json:"description"`
	Status      string `json:"status"` // open, resolved
	CreatedAt   string `json:"created_at"`
}

// Contribution represents a user-submitted link
type Contribution struct {
	ID         int    `json:"id"`
	CourseName string `json:"course_name"`
	LinkURL    string `json:"link_url"`
	LinkType   string `json:"link_type"`
	Note       string `json:"note"`
	Status     string `json:"status"` // pending, approved
	CreatedAt  string `json:"created_at"`
}

// Feedback represents user feedback
type Feedback struct {
	ID        int    `json:"id"`
	Category  string `json:"category"`
	Rating    int    `json:"rating"`
	Message   string `json:"message"`
	Status    string `json:"status"` // new, read
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
	ID        int    `json:"id"`
	LinkID    int    `json:"link_id"`
	ClickedAt string `json:"clicked_at"`
}

// ContentResponse is the big JSON object we send to the frontend.
type ContentResponse struct {
	Programs      []Program      `json:"programs"`
	Years         []Year         `json:"years"`
	Semesters     []Semester     `json:"semesters"`
	Courses       []Course       `json:"courses"`
	Links         []Link         `json:"links"`
	ExtraSections []ExtraSection `json:"extra_sections"`
	ExtraLinks    []Link         `json:"extra_links"`
}
