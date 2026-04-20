package api

import (
	"encoding/json"
	"log"
	"net/http"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"
)

// HandleHealth checks if the server is healthy.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"message": "The Go backend is alive and healthy!",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleApiRoot provides a simple directory of available endpoints.
func HandleApiRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api" && r.URL.Path != "/api/" {
		http.NotFound(w, r)
		return
	}

	response := map[string]interface{}{
		"message": "Welcome to the Info Links API!",
		"usage":   "This is a Go backend serving JSON data.",
		"endpoints": []map[string]string{
			{"path": "/api/health", "method": "GET", "description": "Check if the server is healthy."},
			{"path": "/api/content", "method": "GET", "description": "Fetch the full navigation tree."},
			{"path": "/api/page_views", "method": "POST", "description": "Track a page view."},
			{"path": "/api/link_clicks", "method": "POST", "description": "Track a link click."},
			{"path": "/api/reports", "method": "POST", "description": "Submit a broken link report."},
			{"path": "/api/contributions", "method": "POST", "description": "Submit a new link contribution."},
			{"path": "/api/feedback", "method": "POST", "description": "Submit general feedback."},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetContent fetches all navigation data.
func HandleGetContent(w http.ResponseWriter, r *http.Request) {
	var resp models.ContentResponse

	// 1. Fetch Programs
	rows, err := database.DB.Query("SELECT id, name, display_order FROM programs ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch programs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Program
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayOrder); err != nil {
			log.Println("Scan error (programs):", err)
			continue
		}
		resp.Programs = append(resp.Programs, p)
	}

	// 2. Fetch Years
	rows, err = database.DB.Query("SELECT id, program_id, name, display_order FROM years ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch years", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var y models.Year
		if err := rows.Scan(&y.ID, &y.ProgramID, &y.Name, &y.DisplayOrder); err != nil {
			log.Println("Scan error (years):", err)
			continue
		}
		resp.Years = append(resp.Years, y)
	}

	// 3. Fetch Semesters
	rows, err = database.DB.Query("SELECT id, year_id, name, display_order FROM semesters ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch semesters", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var s models.Semester
		if err := rows.Scan(&s.ID, &s.YearID, &s.Name, &s.DisplayOrder); err != nil {
			log.Println("Scan error (semesters):", err)
			continue
		}
		resp.Semesters = append(resp.Semesters, s)
	}

	// 4. Fetch Courses
	rows, err = database.DB.Query("SELECT id, semester_id, name, code, is_optional, display_order FROM courses ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch courses: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Course
		if err := rows.Scan(&c.ID, &c.SemesterID, &c.Name, &c.Code, &c.IsOptional, &c.DisplayOrder); err != nil {
			log.Println("Scan error (courses):", err)
			continue
		}
		resp.Courses = append(resp.Courses, c)
	}

	// 5. Fetch Links
	rows, err = database.DB.Query("SELECT id, course_id, type, label, url, note, content_type, display_order FROM links WHERE course_id IS NOT NULL ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch links: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var l models.Link
		if err := rows.Scan(&l.ID, &l.CourseID, &l.Type, &l.Label, &l.URL, &l.Note, &l.ContentType, &l.DisplayOrder); err != nil {
			log.Println("Scan error (links):", err)
			continue
		}
		resp.Links = append(resp.Links, l)
	}

	// 6. Fetch Extra Sections
	rows, err = database.DB.Query("SELECT id, title, icon, display_order FROM extra_sections ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch extra sections: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ex models.ExtraSection
		if err := rows.Scan(&ex.ID, &ex.Title, &ex.Icon, &ex.DisplayOrder); err != nil {
			log.Println("Scan error (extra_sections):", err)
			continue
		}
		resp.ExtraSections = append(resp.ExtraSections, ex)
	}

	// 7. Fetch Extra Links
	rows, err = database.DB.Query("SELECT id, section_id, type, label, url, note, content_type, display_order FROM extra_links ORDER BY display_order ASC")
	if err != nil {
		http.Error(w, "Failed to fetch extra links: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var l models.Link
		if err := rows.Scan(&l.ID, &l.SectionID, &l.Type, &l.Label, &l.URL, &l.Note, &l.ContentType, &l.DisplayOrder); err != nil {
			log.Println("Scan error (extra_links):", err)
			continue
		}
		resp.ExtraLinks = append(resp.ExtraLinks, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func HandlePostPageView(w http.ResponseWriter, r *http.Request) {
	var pv models.PageView
	if err := json.NewDecoder(r.Body).Decode(&pv); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := database.DB.Exec("INSERT INTO page_views (page) VALUES ($1)", pv.Page)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostLinkClick(w http.ResponseWriter, r *http.Request) {
	var lc models.LinkClick
	if err := json.NewDecoder(r.Body).Decode(&lc); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := database.DB.Exec("INSERT INTO link_clicks (link_id) VALUES ($1)", lc.LinkID)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostReport(w http.ResponseWriter, r *http.Request) {
	var rep models.Report
	if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := database.DB.Exec("INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)",
		rep.CourseName, rep.LinkURL, rep.Description)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostContribution(w http.ResponseWriter, r *http.Request) {
	var c models.Contribution
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := database.DB.Exec("INSERT INTO contributions (course_name, link_url, note) VALUES ($1, $2, $3)",
		c.CourseName, c.LinkURL, c.Note)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostFeedback(w http.ResponseWriter, r *http.Request) {
	var f models.Feedback
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// The frontend was updated to send: category, rating, message.
	_, err := database.DB.Exec("INSERT INTO feedback (category, rating, message) VALUES ($1, $2, $3)",
        f.Category, f.Rating, f.Message)
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
