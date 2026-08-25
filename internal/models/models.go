package models

import "fmt"

// User represents a user or a guest.
// PreferedLang and PreferedTheme keep the DB spelling (single r) on purpose so
// the column, the struct field and the JSON key never drift apart. They are
// read-only for now: the API exposes them but has no endpoint to change them.
type User struct {
	ID                int    `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Number            int    `json:"number"`
	IsGuest           bool   `json:"is_guest"`
	Handle            string `json:"handle"`
	FavoriteCourseIDs []int  `json:"favorite_course_ids"`
	CreatedAt         string `json:"created_at"`
	LastSeenAt        string `json:"last_seen_at"`
	PreferedLang      string `json:"prefered_lang"`
	PreferedTheme     string `json:"prefered_theme"`
}

// FavoriteEvent represents the action done by the user
type FavoriteEvent struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CourseID  int    `json:"course_id"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
}

// UserListItem is one row of the admin students list, with activity counters.
type UserListItem struct {
	ID         int    `json:"id"`
	Handle     string `json:"handle"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Number     int    `json:"number"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
	VisitCount int    `json:"visit_count"`
	ClickCount int    `json:"click_count"`
}

// UserActivityEvent is one entry of a student activity timeline.
type UserActivityEvent struct {
	Type       string `json:"type"` // visit, link_click, report, contribution, feedback, favorite_added, favorite_removed
	At         string `json:"at"`
	Summary    string `json:"summary"`
	RefID      int    `json:"ref_id"`
	DeviceType string `json:"device_type,omitempty"` // phone or laptop; only set on visits
}

// UserDetail is a student profile plus a page of their activity timeline.
type UserDetail struct {
	User           User                `json:"user"`
	LastDeviceType string              `json:"last_device_type,omitempty"`
	Timeline       []UserActivityEvent `json:"timeline"`
}

// AnalyticsSummary holds the server-side aggregated usage metrics for admins.
type AnalyticsSummary struct {
	TotalStudents           int               `json:"total_students"`
	StudentsGained7d        int               `json:"students_gained_7d"`
	StudentsGained30d       int               `json:"students_gained_30d"`
	StudentsGained90d       int               `json:"students_gained_90d"`
	ActiveToday             int               `json:"active_today"`
	ClicksToday             int               `json:"clicks_today"`
	DevicesToday            DeviceSplit       `json:"devices_today"`
	DailyUniqueVisits       []DailyUniqueDay  `json:"daily_unique_visits"`
	DailyRoster             []DailyRosterDay  `json:"daily_roster"`
	TopLinks                []LinkClickCount  `json:"top_links"`
	TopLinksToday           []LinkClickCount  `json:"top_links_today"`
	TopUsers                []UserClickCount  `json:"top_users"`
	VisitorsToday           VisitorsTodayPage `json:"visitors_today"`
	ActiveInRange           int               `json:"active_in_range"`
	ActiveRegisteredInRange int               `json:"active_registered_in_range"`
	ClicksInRange           int               `json:"clicks_in_range"`
	ClickersInRange         int               `json:"clickers_in_range"`
	ClicksPerActive         float64           `json:"clicks_per_active"`
	PrevActiveInRange       int               `json:"prev_active_in_range"`
	PrevClicksInRange       int               `json:"prev_clicks_in_range"`
	PrevStudentsGained      int               `json:"prev_students_gained"`
	DevicesInRange          DeviceSplit       `json:"devices_in_range"`
	ReturningInRange        int               `json:"returning_in_range"`
	NewInRange              int               `json:"new_in_range"`
	Funnel                  SignupFunnel      `json:"funnel"`
	Inbox                   AnalyticsInbox    `json:"inbox"`
	Browse                  BrowseDepth       `json:"browse"`
	TopCourses              []CourseDemand    `json:"top_courses"`
	ZeroClickCourses        []CourseDemand    `json:"zero_click_courses"`
	ZeroClickLinks          []DeadLink        `json:"zero_click_links"`
	TopFavorites            []CourseDemand    `json:"top_favorites"`
	Heatmap                 []HeatmapCell     `json:"heatmap"`
	SearchTerms             []SearchTermCount `json:"search_terms"`
}

// DeviceSplit counts unique students by coarse device class.
type DeviceSplit struct {
	Phone  int `json:"phone"`
	Laptop int `json:"laptop"`
	Both   int `json:"both"`
}

// SignupFunnel is guest creation vs signup in the selected range.
type SignupFunnel struct {
	Arrivals   int `json:"arrivals"`
	SignedUp   int `json:"signed_up"`
	StillGuest int `json:"still_guest"`
	GuestsOpen int `json:"guests_open"`
}

// AnalyticsInbox is open admin work, not scoped to the chart range.
type AnalyticsInbox struct {
	Reports       int `json:"reports"`
	Contributions int `json:"contributions"`
	Feedback      int `json:"feedback"`
}

// BrowseDepth is unique students who reached each mobile/desktop picker step.
type BrowseDepth struct {
	ReachedYear int `json:"reached_year"`
	ReachedList int `json:"reached_list"`
}

// CourseDemand is a course ranked by clicks or stars.
type CourseDemand struct {
	CourseID    int    `json:"course_id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Count       int    `json:"count"`
	ProgramName string `json:"program_name"`
}

// DeadLink is a resource with no clicks in the selected range.
type DeadLink struct {
	Kind        string `json:"kind"` // link or extra_link
	ID          int    `json:"id"`
	Label       string `json:"label"`
	CourseName  string `json:"course_name"`
	ProgramName string `json:"program_name"`
}

// HeatmapCell is activity (visits + clicks) for one weekday hour.
type HeatmapCell struct {
	Dow   int `json:"dow"` // 0 = Sunday, matching Postgres EXTRACT(DOW)
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// SearchTermCount is how often a search string was submitted in range.
type SearchTermCount struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

// DailyRosterDay is the cumulative registered student count at end of day.
type DailyRosterDay struct {
	Day   string `json:"day"`
	Total int    `json:"total"`
}

// VisitorsTodayPage is one page of today's visitors with a pager hint.
type VisitorsTodayPage struct {
	Visitors []UserClickCount `json:"visitors"`
	HasMore  bool             `json:"has_more"`
}

// DailyUniqueDay counts the distinct users who visited on a given day.
type DailyUniqueDay struct {
	Day   string `json:"day"`
	Users int    `json:"users"`
}

// LinkClickCount counts clicks for a link or an extra link.
type LinkClickCount struct {
	LinkID      *int `json:"link_id"`
	ExtraLinkID *int `json:"extra_link_id"`
	Clicks      int  `json:"clicks"`
}

// UserClickCount counts link clicks for one student.
type UserClickCount struct {
	UserID int    `json:"user_id"`
	Handle string `json:"handle"`
	Clicks int    `json:"clicks"`
}

// UserHandle builds the display handle of a student, e.g. mohamad_hassan_55.
// Guests have no name yet, so they fall back to guest_<id>.
func UserHandle(firstName, lastName string, number, id int) string {
	if firstName == "" || lastName == "" || number == 0 {
		return fmt.Sprintf("guest_%d", id)
	}
	return fmt.Sprintf("%s_%s_%d", firstName, lastName, number)
}

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

// Course is a canonical catalog row. SemesterID and PlacementID are the
// offering: the same id can appear under several programs via placements.
type Course struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	IsOptional   bool   `json:"is_optional"`
	SemesterID   int    `json:"semester_id"`
	PlacementID  int    `json:"placement_id,omitempty"`
	DisplayOrder int    `json:"display_order"`
}

// CoursePatch represents an updating course
type CoursePatch struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	IsOptional  *bool   `json:"is_optional"`
	SemesterID  *int    `json:"semester_id"`
	PlacementID *int    `json:"placement_id"`
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
	UserID      int    `json:"user_id,omitempty"`
	Status      string `json:"status"` // open, resolved, rejected
	LinkURL     string `json:"link_url"`
	CreatedAt   string `json:"created_at"`
	CourseName  string `json:"course_name"`
	Description string `json:"description"`
}

// Contribution represents a user-submitted link
type Contribution struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id,omitempty"`
	Note       string `json:"note"`
	Status     string `json:"status"` // pending, approved, rejected
	LinkURL    string `json:"link_url"`
	LinkType   string `json:"link_type,omitempty"` // accepted on POST; stored in note as [Type:...]
	CreatedAt  string `json:"created_at"`
	CourseName string `json:"course_name"`
}

// Feedback represents user feedback
type Feedback struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id,omitempty"`
	Rating    int    `json:"rating"`
	Status    string `json:"status"` // new, read, rejected
	Message   string `json:"message"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// PageView tracks site visits
type PageView struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id,omitempty"`
	Page       string `json:"page"`
	VisitedAt  string `json:"visited_at"`
	DeviceType string `json:"device_type,omitempty"`
}

// LinkClick tracks clicks on specific links
type LinkClick struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id,omitempty"`
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
