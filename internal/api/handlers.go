package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"infolinks-backend/internal/database"
	"infolinks-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

const maxBodyBytes = 1 << 20

func SetJWTSecret(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	jwtSecret = []byte(secret)
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

func parsePaginationParams(r *http.Request, defaultLimit int) (limit int, offset int, q string) {
	limit = defaultLimit
	if limit <= 0 {
		limit = 25
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100
			}
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	q = strings.TrimSpace(r.URL.Query().Get("q"))
	return
}

// RequireAdmin middleware verifies the JWT token in the Authorization header.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: No token provided")
			return
		}

		// Handle "Bearer <token>" format
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid token claims")
			return
		}
		adminClaim, ok := claims["admin"].(bool)
		if !ok || !adminClaim {
			writeJSONError(w, http.StatusForbidden, "Forbidden: Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// HandleLogin authenticates an admin and returns a JWT.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSONBody(w, r, &creds) {
		return
	}

	// For simplicity, we compare against an environment variable.
	// In a full multi-user system, we would query the users table.
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if creds.Email != adminEmail || creds.Password != adminPass {
		writeJSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(), // 1 week
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

// HandleHealth checks if the server is healthy.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"message": "The Go backend is alive and healthy!",
	}
	writeJSON(w, http.StatusOK, response)
}

// HandleApiRoot provides a simple directory of available endpoints.
func HandleApiRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Welcome to the Info Links API!",
		"usage":   "This is a Go backend serving JSON data.",
		"public_endpoints": []map[string]string{
			{"path": "/api/health", "method": "GET", "description": "Check if the server is healthy."},
			{"path": "/api/content", "method": "GET", "description": "Fetch the full navigation tree."},
			{"path": "/api/auth/login", "method": "POST", "description": "Admin login (returns JWT token)."},
			{"path": "/api/feedback", "method": "POST", "description": "Submit user feedback."},
			{"path": "/api/reports", "method": "POST", "description": "Submit a course/link report."},
		},
		"admin_endpoints": []map[string]string{
			{"path": "/api/admin/courses", "method": "POST/PATCH/DELETE", "description": "Manage courses."},
			{"path": "/api/admin/links", "method": "POST/PATCH/DELETE", "description": "Manage links."},
			{"path": "/api/admin/reports", "method": "GET/PATCH/DELETE", "description": "Manage user reports."},
			{"path": "/api/admin/feedback", "method": "GET/PATCH/DELETE", "description": "Manage feedback."},
			{"path": "/api/admin/page_views", "method": "GET", "description": "View analytics (page views)."},
			{"path": "/api/admin/link_clicks", "method": "GET", "description": "View analytics (link clicks)."},
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleGetContent fetches all navigation data using a single optimized query.
func HandleGetContent(w http.ResponseWriter, r *http.Request) {
	query := `
		WITH content AS (
			SELECT 
				(SELECT COALESCE(json_agg(p ORDER BY display_order ASC), '[]') FROM programs p) as programs,
				(SELECT COALESCE(json_agg(y ORDER BY display_order ASC), '[]') FROM years y) as years,
				(SELECT COALESCE(json_agg(s ORDER BY display_order ASC), '[]') FROM semesters s) as semesters,
				(SELECT COALESCE(json_agg(c ORDER BY display_order ASC), '[]') FROM courses c) as courses,
				(SELECT COALESCE(json_agg(l ORDER BY display_order ASC), '[]') FROM links l WHERE course_id IS NOT NULL) as links,
				(SELECT COALESCE(json_agg(ex ORDER BY display_order ASC), '[]') FROM extra_sections ex) as extra_sections,
				(SELECT COALESCE(json_agg(el ORDER BY display_order ASC), '[]') FROM extra_links el) as extra_links
		)
		SELECT json_build_object(
			'programs', programs,
			'years', years,
			'semesters', semesters,
			'courses', courses,
			'links', links,
			'extra_sections', extra_sections,
			'extra_links', extra_links
		) FROM content;
	`

	var result string
	err := database.DB.QueryRow(query).Scan(&result)
	if err != nil {
		log.Println("Database error in HandleGetContent:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}

func HandlePostPageView(w http.ResponseWriter, r *http.Request) {
	var pv models.PageView
	if !decodeJSONBody(w, r, &pv) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO page_views (page) VALUES ($1)", pv.Page)
	if err != nil {
		log.Println("DB error:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostLinkClick(w http.ResponseWriter, r *http.Request) {
	var lc models.LinkClick
	if !decodeJSONBody(w, r, &lc) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO link_clicks (link_id) VALUES ($1)", lc.LinkID)
	if err != nil {
		log.Println("DB error:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostReport(w http.ResponseWriter, r *http.Request) {
	var rep models.Report
	if !decodeJSONBody(w, r, &rep) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)",
		rep.CourseName, rep.LinkURL, rep.Description)
	if err != nil {
		log.Println("DB error:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostContribution(w http.ResponseWriter, r *http.Request) {
	var c models.Contribution
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO contributions (course_name, link_url, note) VALUES ($1, $2, $3)",
		c.CourseName, c.LinkURL, c.Note)
	if err != nil {
		log.Println("DB error:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandlePostFeedback(w http.ResponseWriter, r *http.Request) {
	var f models.Feedback
	if !decodeJSONBody(w, r, &f) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO feedback (category, rating, message) VALUES ($1, $2, $3)",
		f.Category, f.Rating, f.Message)
	if err != nil {
		log.Println("DB error:", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func HandleAdminGetReports(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, course_name, link_url, description, status, created_at FROM reports"
	var args []interface{}
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(course_name ILIKE $%d OR description ILIKE $%d OR link_url ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var reps []models.Report
	for rows.Next() {
		var rep models.Report
		if err := rows.Scan(&rep.ID, &rep.CourseName, &rep.LinkURL, &rep.Description, &rep.Status, &rep.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		reps = append(reps, rep)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, reps)
}

func HandleAdminUpdateReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid report id")
		return
	}
	var body struct{ Status string `json:"status"` }
	if !decodeJSONBody(w, r, &body) {
		return
	}
	res, err := database.DB.Exec("UPDATE reports SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeJSONError(w, http.StatusNotFound, "Report not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminDeleteReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM reports WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminGetFeedback(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, category, rating, message, status, created_at FROM feedback"
	var args []interface{}
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(category ILIKE $%d OR message ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var feed []models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.ID, &f.Category, &f.Rating, &f.Message, &f.Status, &f.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		feed = append(feed, f)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, feed)
}

func HandleAdminPatchFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct{ Status string `json:"status"` }
	if !decodeJSONBody(w, r, &body) {
		return
	}
	_, err := database.DB.Exec("UPDATE feedback SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminDeleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM feedback WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminGetContributions(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	query := "SELECT id, course_name, link_url, note, status, created_at FROM contributions"
	var args []interface{}
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(course_name ILIKE $%d OR link_url ILIKE $%d OR note ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var contribs []models.Contribution
	for rows.Next() {
		var c models.Contribution
		if err := rows.Scan(&c.ID, &c.CourseName, &c.LinkURL, &c.Note, &c.Status, &c.CreatedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		contribs = append(contribs, c)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, contribs)
}

func HandleAdminUpdateContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct{ Status string `json:"status"` }
	if !decodeJSONBody(w, r, &body) {
		return
	}
	_, err := database.DB.Exec("UPDATE contributions SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminDeleteContribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM contributions WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminPostCourse(w http.ResponseWriter, r *http.Request) {
	var c models.Course
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO courses (semester_id, name, code, is_optional, display_order) VALUES ($1, $2, $3, $4, $5)",
		c.SemesterID, c.Name, c.Code, c.IsOptional, c.DisplayOrder)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func HandleAdminPatchCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var c models.Course
	if !decodeJSONBody(w, r, &c) {
		return
	}
	_, err := database.DB.Exec("UPDATE courses SET name = $1, code = $2, semester_id = $3 WHERE id = $4",
		c.Name, c.Code, c.SemesterID, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminDeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM courses WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminPostLink(w http.ResponseWriter, r *http.Request) {
	var l models.Link
	if !decodeJSONBody(w, r, &l) {
		return
	}
	_, err := database.DB.Exec("INSERT INTO links (course_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		l.CourseID, l.Type, l.URL, l.Label, l.Note, l.ContentType, l.DisplayOrder)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func HandleAdminPatchLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var l models.Link
	if !decodeJSONBody(w, r, &l) {
		return
	}
	_, err := database.DB.Exec("UPDATE links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6",
		l.Type, l.URL, l.Label, l.Note, l.ContentType, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminDeleteLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := database.DB.Exec("DELETE FROM links WHERE id = $1", id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdminGetPageViews(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, page, visited_at FROM page_views ORDER BY visited_at DESC")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var views []models.PageView
	for rows.Next() {
		var v models.PageView
		if err := rows.Scan(&v.ID, &v.Page, &v.VisitedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, views)
}

func HandleAdminGetLinkClicks(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, link_id, clicked_at FROM link_clicks ORDER BY clicked_at DESC")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()
	var clicks []models.LinkClick
	for rows.Next() {
		var c models.LinkClick
		if err := rows.Scan(&c.ID, &c.LinkID, &c.ClickedAt); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		clicks = append(clicks, c)
	}
	if err := rows.Err(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, clicks)
}
