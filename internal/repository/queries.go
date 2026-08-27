package repository

// User Queries
const (
	// favorite_course_ids travels as JSON text so it can be decoded without an array driver.
	userColumns = `id, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(number, 0), is_guest, COALESCE(array_to_json(favorite_course_ids)::text, '[]'), created_at, last_seen_at, prefered_lang, prefered_theme`

	insertNewGuestQuery = `INSERT INTO users (is_guest) VALUES (true) RETURNING id`
	insertNewUserQuery  = `INSERT INTO users (first_name,last_name,number,is_guest) VALUES ($1,$2,$3,false) RETURNING ` + userColumns
	claimGuestQuery     = `UPDATE users SET first_name = $1, last_name = $2, number = $3, is_guest = false, last_seen_at = now() WHERE id = $4 AND is_guest = true RETURNING ` + userColumns

	getUserByIDQuery          = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	getUserByCredentialsQuery = `SELECT ` + userColumns + ` FROM users WHERE first_name = $1 AND last_name = $2 AND number = $3 AND is_guest = false`

	// Sign-in cannot claim the guest row (that name already belongs to a student),
	// so activity is moved across and the empty guest is deleted.
	lockGuestForAdoptQuery      = `SELECT id FROM users WHERE id = $1 AND is_guest = true FOR UPDATE`
	reassignPageViewsQuery      = `UPDATE page_views SET user_id = $2 WHERE user_id = $1`
	reassignLinkClicksQuery     = `UPDATE link_clicks SET user_id = $2 WHERE user_id = $1`
	reassignReportsQuery        = `UPDATE reports SET user_id = $2 WHERE user_id = $1`
	reassignContributionsQuery  = `UPDATE contributions SET user_id = $2 WHERE user_id = $1`
	reassignFeedbackQuery       = `UPDATE feedback SET user_id = $2 WHERE user_id = $1`
	reassignFavoriteEventsQuery = `UPDATE favorite_events SET user_id = $2 WHERE user_id = $1`
	reassignSearchEventsQuery   = `UPDATE search_events SET user_id = $2 WHERE user_id = $1`
	reassignBrowseEventsQuery   = `UPDATE browse_events SET user_id = $2 WHERE user_id = $1`
	deleteGuestQuery            = `DELETE FROM users WHERE id = $1 AND is_guest = true`
	// Cascades page_views / search / browse for those guests; registered rows are untouched.
	deleteStaleGuestsQuery = `DELETE FROM users WHERE is_guest = true AND last_seen_at < $1`
	touchLastSeenQuery     = `UPDATE users SET last_seen_at = now() WHERE id = $1`
)

// Favorites Queries
const (
	addFavoriteQuery         = `UPDATE users SET favorite_course_ids = array_append(favorite_course_ids, $2), last_seen_at = now() WHERE id = $1 AND NOT (favorite_course_ids @> ARRAY[$2::integer])`
	removeFavoriteQuery      = `UPDATE users SET favorite_course_ids = array_remove(favorite_course_ids, $2), last_seen_at = now() WHERE id = $1 AND favorite_course_ids @> ARRAY[$2::integer]`
	insertFavoriteEventQuery = `INSERT INTO favorite_events (user_id,course_id,action) VALUES ($1,$2,$3)`
)

// Admin Students Queries
const (
	studentColumns = `u.id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.number, 0), u.created_at, u.last_seen_at,
		       (SELECT COUNT(*) FROM page_views pv WHERE pv.user_id = u.id),
		       (SELECT COUNT(*) FROM link_clicks lc WHERE lc.user_id = u.id)`

	listStudentsBaseQuery  = `SELECT ` + studentColumns + ` FROM users u WHERE u.is_guest = false`
	listStudentsOrderQuery = ` ORDER BY u.first_name ASC, u.last_name ASC LIMIT `

	listStudentsQuery      = listStudentsBaseQuery + listStudentsOrderQuery + `$1 OFFSET $2`
	listStudentsWithQQuery = listStudentsBaseQuery + ` AND (u.first_name ILIKE $1 OR u.last_name ILIKE $1)` + listStudentsOrderQuery + `$2 OFFSET $3`

	// listUserTimelineQuery merges every activity table into one chronological feed.
	// device_type is only stored on page_views; other event types return ''.
	listUserTimelineQuery = `
		SELECT type, at, summary, ref_id, device_type FROM (
			SELECT 'visit' AS type, pv.visited_at AS at,
			       'visited ' || COALESCE(pv.page, 'home')
			           || CASE WHEN pv.device_type IN ('phone', 'laptop')
			                   THEN ' from ' || pv.device_type ELSE '' END AS summary,
			       pv.id AS ref_id,
			       COALESCE(pv.device_type, '') AS device_type
			FROM page_views pv WHERE pv.user_id = $1
			UNION ALL
			SELECT 'link_click', lc.clicked_at,
			       'opened ' || COALESCE(l.label, el.label, 'a link')
			           || COALESCE(' in ' || co.name, ' in ' || es.title, ''),
			       lc.id, ''
			FROM link_clicks lc
			LEFT JOIN links l ON l.id = lc.link_id
			LEFT JOIN courses co ON co.id = l.course_id
			LEFT JOIN extra_links el ON el.id = lc.extra_link_id
			LEFT JOIN extra_sections es ON es.id = el.section_id
			WHERE lc.user_id = $1
			UNION ALL
			SELECT 'report', r.created_at,
			       'reported a link in ' || r.course_name, r.id, ''
			FROM reports r WHERE r.user_id = $1
			UNION ALL
			SELECT 'contribution', c.created_at,
			       'suggested a link for ' || c.course_name, c.id, ''
			FROM contributions c WHERE c.user_id = $1
			UNION ALL
			SELECT 'feedback', f.created_at,
			       'sent ' || f.category || ' feedback rated ' || f.rating || '/5', f.id, ''
			FROM feedback f WHERE f.user_id = $1
			UNION ALL
			SELECT CASE WHEN fe.action = 'added' THEN 'favorite_added' ELSE 'favorite_removed' END,
			       fe.created_at,
			       CASE WHEN fe.action = 'added'
			            THEN 'added ' || co.name || ' to favorites'
			            ELSE 'removed ' || co.name || ' from favorites' END,
			       fe.id, ''
			FROM favorite_events fe
			JOIN courses co ON co.id = fe.course_id
			WHERE fe.user_id = $1
		) timeline
		ORDER BY at DESC
		LIMIT $2 OFFSET $3`

	getLastDeviceTypeQuery = `SELECT device_type FROM page_views WHERE user_id = $1 AND device_type IS NOT NULL ORDER BY visited_at DESC LIMIT 1`
)

// Admin Analytics Queries
const (
	analyticsCountsQuery = `
		SELECT
			(SELECT COUNT(*) FROM users WHERE is_guest = false),
			(SELECT COUNT(*) FROM users WHERE is_guest = false AND created_at >= now() - interval '7 days'),
			(SELECT COUNT(*) FROM users WHERE is_guest = false AND created_at >= now() - interval '30 days'),
			(SELECT COUNT(*) FROM users WHERE is_guest = false AND created_at >= now() - interval '90 days'),
			(SELECT COUNT(DISTINCT uid) FROM (
				SELECT user_id AS uid FROM page_views
				WHERE user_id IS NOT NULL AND visited_at >= date_trunc('day', now())
				UNION
				SELECT user_id AS uid FROM link_clicks
				WHERE user_id IS NOT NULL AND clicked_at >= date_trunc('day', now())
			) active_today),
			(SELECT COUNT(*) FROM link_clicks WHERE clicked_at >= date_trunc('day', now())),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= date_trunc('day', now()) AND device_type = 'phone'),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= date_trunc('day', now()) AND device_type = 'laptop'),
			(SELECT COUNT(*) FROM (
				SELECT user_id FROM page_views
				WHERE user_id IS NOT NULL AND visited_at >= date_trunc('day', now()) AND device_type IS NOT NULL
				GROUP BY user_id HAVING COUNT(DISTINCT device_type) > 1
			) both_today),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM link_clicks WHERE clicked_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(DISTINCT user_id) FROM link_clicks WHERE user_id IS NOT NULL AND clicked_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1 * 2) AND visited_at < now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM link_clicks WHERE clicked_at >= now() - make_interval(days => $1 * 2) AND clicked_at < now() - make_interval(days => $1)),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1) AND device_type = 'phone'),
			(SELECT COUNT(DISTINCT user_id) FROM page_views WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1) AND device_type = 'laptop'),
			(SELECT COUNT(*) FROM (
				SELECT user_id FROM page_views
				WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1) AND device_type IS NOT NULL
				GROUP BY user_id HAVING COUNT(DISTINCT device_type) > 1
			) both_range),
			(SELECT COUNT(DISTINCT pv.user_id) FROM page_views pv
				WHERE pv.user_id IS NOT NULL AND pv.visited_at >= now() - make_interval(days => $1)
				AND EXISTS (SELECT 1 FROM page_views older WHERE older.user_id = pv.user_id AND older.visited_at < now() - make_interval(days => $1))),
			(SELECT COUNT(DISTINCT pv.user_id) FROM page_views pv
				WHERE pv.user_id IS NOT NULL AND pv.visited_at >= now() - make_interval(days => $1)
				AND NOT EXISTS (SELECT 1 FROM page_views older WHERE older.user_id = pv.user_id AND older.visited_at < now() - make_interval(days => $1))),
			(SELECT COUNT(*) FROM users WHERE created_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM users WHERE is_guest = false AND created_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM users WHERE is_guest = false AND created_at >= now() - make_interval(days => $1 * 2) AND created_at < now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM users WHERE is_guest = true AND created_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(*) FROM users WHERE is_guest = true),
			(SELECT COUNT(*) FROM reports WHERE status = 'open'),
			(SELECT COUNT(*) FROM contributions WHERE status = 'pending'),
			(SELECT COUNT(*) FROM feedback WHERE status = 'new'),
			(SELECT COUNT(DISTINCT user_id) FROM browse_events WHERE step = 'year' AND created_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(DISTINCT user_id) FROM browse_events WHERE step = 'list' AND created_at >= now() - make_interval(days => $1)),
			(SELECT COUNT(DISTINCT u.id) FROM users u
				WHERE u.is_guest = false
				AND (
					EXISTS (
						SELECT 1 FROM page_views pv
						WHERE pv.user_id = u.id AND pv.visited_at >= now() - make_interval(days => $1)
					)
					OR EXISTS (
						SELECT 1 FROM link_clicks lc
						WHERE lc.user_id = u.id AND lc.clicked_at >= now() - make_interval(days => $1)
					)
				))`

	analyticsDailyUniqueVisitsQuery = `
		SELECT to_char(visited_at, 'YYYY-MM-DD') AS day, COUNT(DISTINCT user_id)
		FROM page_views
		WHERE user_id IS NOT NULL AND visited_at >= now() - make_interval(days => $1)
		GROUP BY day
		ORDER BY day ASC`

	analyticsTopLinksQuery = `
		SELECT link_id, extra_link_id, COUNT(*) AS clicks
		FROM link_clicks
		WHERE clicked_at >= now() - make_interval(days => $1)
		GROUP BY link_id, extra_link_id
		ORDER BY clicks DESC
		LIMIT 50`

	analyticsTopUsersQuery = `
		SELECT u.id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.number, 0), COUNT(lc.id) AS clicks
		FROM link_clicks lc
		JOIN users u ON u.id = lc.user_id
		WHERE lc.clicked_at >= now() - make_interval(days => $1)
		GROUP BY u.id, u.first_name, u.last_name, u.number
		ORDER BY clicks DESC, u.first_name ASC
		LIMIT 50`

	analyticsTopLinksTodayQuery = `
		SELECT link_id, extra_link_id, COUNT(*) AS clicks
		FROM link_clicks
		WHERE clicked_at >= date_trunc('day', now())
		GROUP BY link_id, extra_link_id
		ORDER BY clicks DESC
		LIMIT 50`

	analyticsDailyRosterQuery = `
		WITH days AS (
			SELECT generate_series(
				date_trunc('day', now()) - make_interval(days => $1 - 1),
				date_trunc('day', now()),
				interval '1 day'
			) AS day
		)
		SELECT to_char(d.day, 'YYYY-MM-DD'),
		       COUNT(u.id)
		FROM days d
		LEFT JOIN users u ON u.is_guest = false AND u.created_at < d.day + interval '1 day'
		GROUP BY d.day
		ORDER BY d.day ASC`

	// A visitor is anyone with a page view or a link click today. Link opens are
	// gated behind signup, so a click without a page_views row still means the
	// person was on the site (e.g. visit POST failed or session was re-bootstrapped).
	analyticsVisitorsTodayByClicksQuery = `
		SELECT u.id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.number, 0),
		       (SELECT COUNT(*) FROM link_clicks lc WHERE lc.user_id = u.id AND lc.clicked_at >= date_trunc('day', now())) AS clicks
		FROM users u
		WHERE EXISTS (SELECT 1 FROM page_views pv WHERE pv.user_id = u.id AND pv.visited_at >= date_trunc('day', now()))
		   OR EXISTS (SELECT 1 FROM link_clicks lc WHERE lc.user_id = u.id AND lc.clicked_at >= date_trunc('day', now()))
		ORDER BY clicks DESC, u.first_name ASC, u.last_name ASC, u.id ASC
		LIMIT $1 OFFSET $2`

	analyticsVisitorsTodayByNameQuery = `
		SELECT u.id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), COALESCE(u.number, 0),
		       (SELECT COUNT(*) FROM link_clicks lc WHERE lc.user_id = u.id AND lc.clicked_at >= date_trunc('day', now())) AS clicks
		FROM users u
		WHERE EXISTS (SELECT 1 FROM page_views pv WHERE pv.user_id = u.id AND pv.visited_at >= date_trunc('day', now()))
		   OR EXISTS (SELECT 1 FROM link_clicks lc WHERE lc.user_id = u.id AND lc.clicked_at >= date_trunc('day', now()))
		ORDER BY u.first_name ASC, u.last_name ASC, u.id ASC
		LIMIT $1 OFFSET $2`

	analyticsTopCoursesQuery = `
		SELECT c.id, c.name, c.code, COUNT(*)::int, COALESCE((
			SELECT string_agg(DISTINCT pr.name, ' · ' ORDER BY pr.name)
			FROM course_placements pl
			JOIN semesters s ON s.id = pl.semester_id
			JOIN years y ON y.id = s.year_id
			JOIN programs pr ON pr.id = y.program_id
			WHERE pl.course_id = c.id
		), '')
		FROM link_clicks lc
		JOIN links l ON l.id = lc.link_id
		JOIN courses c ON c.id = l.course_id
		WHERE lc.clicked_at >= now() - make_interval(days => $1)
		GROUP BY c.id, c.name, c.code
		ORDER BY COUNT(*) DESC, c.name ASC
		LIMIT 50`

	analyticsZeroClickCoursesQuery = `
		SELECT c.id, c.name, c.code, 0, COALESCE((
			SELECT string_agg(DISTINCT pr.name, ' · ' ORDER BY pr.name)
			FROM course_placements pl
			JOIN semesters s ON s.id = pl.semester_id
			JOIN years y ON y.id = s.year_id
			JOIN programs pr ON pr.id = y.program_id
			WHERE pl.course_id = c.id
		), '')
		FROM courses c
		WHERE EXISTS (SELECT 1 FROM links l WHERE l.course_id = c.id)
		  AND NOT EXISTS (
			SELECT 1 FROM links l
			JOIN link_clicks lc ON lc.link_id = l.id AND lc.clicked_at >= now() - make_interval(days => $1)
			WHERE l.course_id = c.id
		  )
		ORDER BY c.name ASC
		LIMIT 50`

	analyticsZeroClickLinksQuery = `
		SELECT kind, id, label, course_name, program_name FROM (
			SELECT 'link'::text AS kind, l.id, COALESCE(l.label, 'Link') AS label, c.name AS course_name,
			       COALESCE((
				SELECT string_agg(DISTINCT pr.name, ' · ' ORDER BY pr.name)
				FROM course_placements pl
				JOIN semesters s ON s.id = pl.semester_id
				JOIN years y ON y.id = s.year_id
				JOIN programs pr ON pr.id = y.program_id
				WHERE pl.course_id = c.id
			       ), '') AS program_name
			FROM links l
			JOIN courses c ON c.id = l.course_id
			WHERE NOT EXISTS (
				SELECT 1 FROM link_clicks lc
				WHERE lc.link_id = l.id AND lc.clicked_at >= now() - make_interval(days => $1)
			)
			UNION ALL
			SELECT 'extra_link', el.id, COALESCE(el.label, 'Link'), es.title, ''
			FROM extra_links el
			JOIN extra_sections es ON es.id = el.section_id
			WHERE NOT EXISTS (
				SELECT 1 FROM link_clicks lc
				WHERE lc.extra_link_id = el.id AND lc.clicked_at >= now() - make_interval(days => $1)
			)
		) gaps
		ORDER BY course_name ASC, label ASC
		LIMIT 50`

	analyticsTopFavoritesQuery = `
		SELECT c.id, c.name, c.code, COUNT(*)::int, COALESCE((
			SELECT string_agg(DISTINCT pr.name, ' · ' ORDER BY pr.name)
			FROM course_placements pl
			JOIN semesters s ON s.id = pl.semester_id
			JOIN years y ON y.id = s.year_id
			JOIN programs pr ON pr.id = y.program_id
			WHERE pl.course_id = c.id
		), '')
		FROM users u
		CROSS JOIN LATERAL unnest(u.favorite_course_ids) AS cid
		JOIN courses c ON c.id = cid
		WHERE u.is_guest = false
		GROUP BY c.id, c.name, c.code
		ORDER BY COUNT(*) DESC, c.name ASC
		LIMIT 50`

	analyticsHeatmapQuery = `
		SELECT EXTRACT(DOW FROM ts)::int, EXTRACT(HOUR FROM ts)::int, COUNT(*)::int
		FROM (
			SELECT clicked_at AS ts FROM link_clicks WHERE clicked_at >= now() - make_interval(days => $1)
			UNION ALL
			SELECT visited_at FROM page_views WHERE visited_at >= now() - make_interval(days => $1)
		) activity
		GROUP BY 1, 2`

	analyticsSearchTermsQuery = `
		SELECT query, COUNT(*)::int
		FROM search_events
		WHERE created_at >= now() - make_interval(days => $1)
		GROUP BY query
		ORDER BY COUNT(*) DESC, query ASC
		LIMIT 50`

	insertSearchEventQuery = `INSERT INTO search_events (user_id, query) VALUES ($1, $2)`
	insertBrowseEventQuery = `INSERT INTO browse_events (user_id, step) VALUES ($1, $2)`
)

// Page Views Queries
const (
	// One statement keeps the visit row and the last_seen_at touch atomic.
	insertPageViewQuery = `WITH visit AS (INSERT INTO page_views (page,user_id,device_type) VALUES ($1,$2,$3)) UPDATE users SET last_seen_at = now() WHERE id = $2`
	GetPageViewQuery    = `SELECT id, page, visited_at FROM page_views ORDER BY visited_at DESC`
)

// Link Clicks Queries
const (
	insertLinkClickQuery = `WITH click AS (INSERT INTO link_clicks (link_id,extra_link_id,user_id) VALUES ($1,$2,$3)) UPDATE users SET last_seen_at = now() WHERE id = $3`
	GetLinkClickQuery    = `SELECT id, link_id, extra_link_id, clicked_at FROM link_clicks ORDER BY clicked_at DESC`
)

// Courses Queries
const (
	getCourseByIDQuery      = `SELECT id, name, code, is_optional FROM courses WHERE id = $1`
	deleteCourseQuery       = `DELETE FROM courses WHERE id = $1`
	updateCourseQuery       = `UPDATE courses SET name = $1, code = $2, is_optional = $3 WHERE id = $4`
	findCourseIDByCodeQuery = `
		SELECT id FROM courses
		WHERE lower(trim(code)) = lower(trim($1))
		LIMIT 1`
	insertCanonicalCourseQuery = `INSERT INTO courses (name, code, is_optional) VALUES ($1, $2, $3) RETURNING id`
	insertCoursePlacementQuery = `
		INSERT INTO course_placements (course_id, semester_id, display_order)
		VALUES ($1, $2, $3)`
	updateCoursePlacementQuery = `
		UPDATE course_placements SET semester_id = $1, display_order = $2
		WHERE id = $3 AND course_id = $4`
	deleteCoursePlacementQuery = `DELETE FROM course_placements WHERE id = $1 AND course_id = $2`
	deleteOrphanCourseQuery    = `
		DELETE FROM courses c
		WHERE c.id = $1
		  AND NOT EXISTS (SELECT 1 FROM course_placements p WHERE p.course_id = c.id)`
)

// Links Queries
const (
	deleteLinkQuery = `DELETE FROM links WHERE id = $1`
	updateLinkQuery = `UPDATE links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6`
	insertLinkQuery = `INSERT INTO links (course_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

// Extra sections queries
const (
	listExtraSectionsQuery         = `SELECT id, title, icon, display_order FROM extra_sections ORDER BY display_order ASC`
	insertExtraSectionQuery        = `INSERT INTO extra_sections (title, icon, display_order) VALUES ($1, $2, $3)`
	updateExtraSectionQuery        = `UPDATE extra_sections SET title = $1, icon = $2 WHERE id = $3`
	deleteExtraSectionQuery        = `DELETE FROM extra_sections WHERE id = $1`
	deleteExtraLinksBySectionQuery = `DELETE FROM extra_links WHERE section_id = $1`
)

// Extra links queries
const (
	listExtraLinksQuery  = `SELECT id, section_id, type, url, label, note, content_type, display_order FROM extra_links ORDER BY display_order ASC`
	insertExtraLinkQuery = `INSERT INTO extra_links (section_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	updateExtraLinkQuery = `UPDATE extra_links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6`
	deleteExtraLinkQuery = `DELETE FROM extra_links WHERE id = $1`
)

// Feedbacks Queries
const (
	deleteFeedbackQuery = `DELETE FROM feedback WHERE id = $1`
	updateFeedbackQuery = `UPDATE feedback SET status = $1 WHERE id = $2`
	insertFeedbackQuery = `INSERT INTO feedback (category, rating, message, user_id) VALUES ($1, $2, $3, $4)`

	listFeedbackBaseQuery        = `SELECT id, category, rating, message, status, created_at, user_id FROM feedback`
	listFeedbackNoFilterQuery    = listFeedbackBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listFeedbackWithStatusQuery  = listFeedbackBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithQQuery       = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithQStatusQuery = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// Reports Queries
const (
	deleteReportQuery = `DELETE FROM reports WHERE id = $1`
	updateReportQuery = `UPDATE reports SET status = $1 WHERE id = $2`
	insertReportQuery = `INSERT INTO reports (course_name, link_url, description, user_id) VALUES ($1, $2, $3, $4)`

	listReportsBaseQuery        = `SELECT id, course_name, link_url, description, status, created_at, user_id FROM reports`
	listReportsNoFilterQuery    = listReportsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listReportsWithStatusQuery  = listReportsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithQQuery       = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithQStatusQuery = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// Contributions Queries
const (
	deleteContributionQuery = `DELETE FROM contributions WHERE id = $1`
	updateContributionQuery = `UPDATE contributions SET status = $1 WHERE id = $2`
	insertContributionQuery = `INSERT INTO contributions (course_name, link_url, note, user_id) VALUES ($1, $2, $3, $4)`

	listContributionsBaseQuery        = `SELECT id, course_name, link_url, note, status, created_at, user_id FROM contributions`
	listContributionsNoFilterQuery    = listContributionsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listContributionsWithStatusQuery  = listContributionsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithQQuery       = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR note ILIKE $1 ) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithQStatusQuery = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR note ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

// SEO Queries
const (
	getSEOCoursePlacementsQuery = `
		SELECT c.id, c.name, c.code, c.is_optional,
		       p.id, p.name, y.name, s.name
		FROM courses c
		JOIN course_placements pl ON pl.course_id = c.id
		JOIN semesters s ON pl.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		JOIN programs p ON y.program_id = p.id
		WHERE LOWER(TRIM(c.code)) = LOWER(TRIM($1))
		ORDER BY p.display_order, y.display_order, s.display_order, pl.display_order`

	listSEOLinksByCourseIDsQuery = `
		SELECT l.id, l.label, l.url, COALESCE(l.note, ''),
		       COALESCE(l.content_type, ''), COALESCE(l.type, '')
		FROM links l
		WHERE l.course_id IN (%s)
		ORDER BY l.display_order ASC`

	listSEOCourseCodesForSitemapQuery = `
		SELECT DISTINCT LOWER(TRIM(code))
		FROM courses
		WHERE code IS NOT NULL AND TRIM(code) <> ''
		ORDER BY 1`

	listSEOProgramsQuery = `SELECT id, name FROM programs ORDER BY display_order ASC`

	listSEOCoursesIndexQuery = `
		SELECT DISTINCT ON (LOWER(TRIM(c.code)))
		       LOWER(TRIM(c.code)), c.name, p.name
		FROM courses c
		JOIN course_placements pl ON pl.course_id = c.id
		JOIN semesters s ON pl.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		JOIN programs p ON y.program_id = p.id
		WHERE c.code IS NOT NULL AND TRIM(c.code) <> ''
		ORDER BY LOWER(TRIM(c.code)), p.display_order, pl.display_order`

	listSEOProgramCoursesQuery = `
		SELECT DISTINCT ON (LOWER(TRIM(c.code))) LOWER(TRIM(c.code)), c.name
		FROM courses c
		JOIN course_placements pl ON pl.course_id = c.id
		JOIN semesters s ON pl.semester_id = s.id
		JOIN years y ON s.year_id = y.id
		WHERE y.program_id = $1 AND c.code IS NOT NULL AND TRIM(c.code) <> ''
		ORDER BY LOWER(TRIM(c.code)), pl.display_order`
)

// Contents Queries
const (
	getContentQuery = `
	WITH content AS (
		SELECT
			(SELECT COALESCE(json_agg(y ORDER BY display_order ASC), '[]') FROM years y) as years,
			(SELECT COALESCE(json_agg(c ORDER BY c.display_order ASC), '[]') FROM (
				SELECT c.id, pl.id AS placement_id, pl.semester_id, c.name, c.code, c.is_optional, pl.display_order
				FROM course_placements pl
				JOIN courses c ON c.id = pl.course_id
			) c) as courses,
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
