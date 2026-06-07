package repository

// Page Views Queries
const (
	insertPageViewQuery = `INSERT INTO page_views (page) VALUES ($1)`
	GetPageViewQuery    = `SELECT id, page, visited_at FROM page_views ORDER BY visited_at DESC`
)

// Link Clicks Queries
const (
	insertLinkClickQuery = `INSERT INTO link_clicks (link_id) VALUES ($1)`
	GetLinkClickQuery    = `SELECT id, link_id, clicked_at FROM link_clicks WHERE link_id IS NOT NULL ORDER BY clicked_at DESC`
)

// Courses Queries
const (
	getCourseByIDQuery = `SELECT id, semester_id, name, code, is_optional, display_order FROM courses WHERE id = $1`
	deleteCourseQuery  = `DELETE FROM courses WHERE id = $1`
	updateCourseQuery  = `UPDATE courses SET name = $1, code = $2, semester_id = $3, is_optional = $4 WHERE id = $5`
	insertCourseQuery  = `INSERT INTO courses (semester_id, name, code, is_optional, display_order) VALUES ($1, $2, $3, $4, $5)`
)

// Links Queries
const (
	deleteLinkQuery = `DELETE FROM links WHERE id = $1`
	updateLinkQuery = `UPDATE links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6`
	insertLinkQuery = `INSERT INTO links (course_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

// Feedbacks Queries
const (
	deleteFeedbackQuery = `DELETE FROM feedback WHERE id = $1`
	updateFeedbackQuery = `UPDATE feedback SET status = $1 WHERE id = $2`
	insertFeedbackQuery = `INSERT INTO feedback (category, rating, message) VALUES ($1, $2, $3)`

	listFeedbackBaseQuery        = `SELECT id, category, rating, message, status, created_at FROM feedback`
	listFeedbackNoFilterQuery    = listFeedbackBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listFeedbackWithStatusQuery  = listFeedbackBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithQQuery       = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithQStatusQuery = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// Reports Queries
const (
	deleteReportQuery = `DELETE FROM reports WHERE id = $1`
	updateReportQuery = `UPDATE reports SET status = $1 WHERE id = $2`
	insertReportQuery = `INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)`

	listReportsBaseQuery        = `SELECT id, course_name, link_url, description, status, created_at FROM reports`
	listReportsNoFilterQuery    = listReportsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listReportsWithStatusQuery  = listReportsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithQQuery       = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithQStatusQuery = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// Contributions Queries
const (
	deleteContributionQuery = `DELETE FROM contributions WHERE id = $1`
	updateContributionQuery = `UPDATE contributions SET status = $1 WHERE id = $2`
	insertContributionQuery = `INSERT INTO contributions (course_name, link_url, note) VALUES ($1, $2, $3)`

	listContributionsBaseQuery        = `SELECT id, course_name, link_url, note, status, created_at FROM contributions`
	listContributionsNoFilterQuery    = listContributionsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listContributionsWithStatusQuery  = listContributionsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithQQuery       = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR note ILIKE $1 ) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithQStatusQuery = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR note ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// Contents Queries
const (
	getContentQuery = `
	WITH content AS (
		SELECT
			(SELECT COALESCE(json_agg(y ORDER BY display_order ASC), '[]') FROM years y) as years,
			(SELECT COALESCE(json_agg(c ORDER BY display_order ASC), '[]') FROM courses c) as courses,
			(SELECT COALESCE(json_agg(p ORDER BY display_order ASC), '[]') FROM programs p) as programs,
			(SELECT COALESCE(json_agg(s ORDER BY display_order ASC), '[]') FROM semesters s) as semesters,
			(SELECT COALESCE(json_agg(el ORDER BY display_order ASC), '[]') FROM extra_links el) as extra_links,
			(SELECT COALESCE(json_agg(ex ORDER BY display_order ASC), '[]') FROM extra_sections ex) as extra_sections,
			(SELECT COALESCE(json_agg(l ORDER BY display_order ASC), '[]') FROM links l WHERE course_id IS NOT NULL) as links
		)
	SELECT json_build_object(
		'years', years,
		'links', links,
		'courses', courses,
		'programs', programs,
		'semesters', semesters,
		'extra_links', extra_links,
		'extra_sections', extra_sections
	) FROM content;
    `
)
