package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

type postgresAnalyticsRepository struct {
	db *sql.DB
}

func NewPostgresAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &postgresAnalyticsRepository{db: db}
}

// GetSummary aggregates usage metrics in Postgres so admins never download raw
// page_views or link_clicks tables.
func (r *postgresAnalyticsRepository) GetSummary(ctx context.Context, params AnalyticsSummaryParams) (models.AnalyticsSummary, error) {
	var summary models.AnalyticsSummary

	if err := r.db.QueryRowContext(ctx, analyticsCountsQuery, params.Days).Scan(
		&summary.TotalStudents,
		&summary.StudentsGained7d,
		&summary.StudentsGained30d,
		&summary.StudentsGained90d,
		&summary.ActiveToday,
		&summary.ClicksToday,
		&summary.DevicesToday.Phone,
		&summary.DevicesToday.Laptop,
		&summary.DevicesToday.Both,
		&summary.ActiveInRange,
		&summary.ClicksInRange,
		&summary.ClickersInRange,
		&summary.PrevActiveInRange,
		&summary.PrevClicksInRange,
		&summary.DevicesInRange.Phone,
		&summary.DevicesInRange.Laptop,
		&summary.DevicesInRange.Both,
		&summary.ReturningInRange,
		&summary.NewInRange,
		&summary.Funnel.Arrivals,
		&summary.Funnel.SignedUp,
		&summary.PrevStudentsGained,
		&summary.Funnel.StillGuest,
		&summary.Funnel.GuestsOpen,
		&summary.Inbox.Reports,
		&summary.Inbox.Contributions,
		&summary.Inbox.Feedback,
		&summary.Browse.ReachedYear,
		&summary.Browse.ReachedList,
		&summary.ActiveRegisteredInRange,
	); err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics counts: %w", err)
	}
	if summary.ActiveInRange > 0 {
		summary.ClicksPerActive = float64(summary.ClicksInRange) / float64(summary.ActiveInRange)
	}

	dailyVisits, err := r.dailyUniqueVisits(ctx, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, err
	}
	summary.DailyUniqueVisits = dailyVisits

	dailyRoster, err := r.dailyRoster(ctx, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, err
	}
	summary.DailyRoster = dailyRoster

	topLinks, err := r.topLinks(ctx, analyticsTopLinksQuery, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, err
	}
	summary.TopLinks = topLinks

	topUsers, err := r.userClicks(ctx, analyticsTopUsersQuery, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics top users: %w", err)
	}
	summary.TopUsers = topUsers

	topLinksToday, err := r.topLinks(ctx, analyticsTopLinksTodayQuery)
	if err != nil {
		return models.AnalyticsSummary{}, err
	}
	summary.TopLinksToday = topLinksToday

	visitorsToday, err := r.visitorsToday(ctx, params)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics visitors today: %w", err)
	}
	summary.VisitorsToday = visitorsToday

	topCourses, err := r.courseDemand(ctx, analyticsTopCoursesQuery, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics top courses: %w", err)
	}
	summary.TopCourses = topCourses

	zeroCourses, err := r.courseDemand(ctx, analyticsZeroClickCoursesQuery, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics zero-click courses: %w", err)
	}
	summary.ZeroClickCourses = zeroCourses

	zeroLinks, err := r.deadLinks(ctx, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics zero-click links: %w", err)
	}
	summary.ZeroClickLinks = zeroLinks

	topFavorites, err := r.courseDemand(ctx, analyticsTopFavoritesQuery)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics top favorites: %w", err)
	}
	summary.TopFavorites = topFavorites

	heatmap, err := r.heatmap(ctx, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics heatmap: %w", err)
	}
	summary.Heatmap = heatmap

	searchTerms, err := r.searchTerms(ctx, params.Days)
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("analytics search terms: %w", err)
	}
	summary.SearchTerms = searchTerms

	return summary, nil
}

func (r *postgresAnalyticsRepository) InsertSearch(ctx context.Context, userID int, query string) error {
	if _, err := r.db.ExecContext(ctx, insertSearchEventQuery, userID, query); err != nil {
		return fmt.Errorf("insert search event: %w", err)
	}
	return nil
}

func (r *postgresAnalyticsRepository) InsertBrowse(ctx context.Context, userID int, step string) error {
	if _, err := r.db.ExecContext(ctx, insertBrowseEventQuery, userID, step); err != nil {
		return fmt.Errorf("insert browse event: %w", err)
	}
	return nil
}

func (r *postgresAnalyticsRepository) dailyUniqueVisits(ctx context.Context, days int) ([]models.DailyUniqueDay, error) {
	rows, err := r.db.QueryContext(ctx, analyticsDailyUniqueVisitsQuery, days)
	if err != nil {
		return nil, fmt.Errorf("analytics daily unique visits query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	visits := []models.DailyUniqueDay{}
	for rows.Next() {
		var v models.DailyUniqueDay
		if err := rows.Scan(&v.Day, &v.Users); err != nil {
			return nil, fmt.Errorf("analytics daily unique visits rows scan: %w", err)
		}
		visits = append(visits, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics daily unique visits rows err: %w", err)
	}
	return visits, nil
}

func (r *postgresAnalyticsRepository) dailyRoster(ctx context.Context, days int) ([]models.DailyRosterDay, error) {
	rows, err := r.db.QueryContext(ctx, analyticsDailyRosterQuery, days)
	if err != nil {
		return nil, fmt.Errorf("analytics daily roster query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	roster := []models.DailyRosterDay{}
	for rows.Next() {
		var d models.DailyRosterDay
		if err := rows.Scan(&d.Day, &d.Total); err != nil {
			return nil, fmt.Errorf("analytics daily roster rows scan: %w", err)
		}
		roster = append(roster, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics daily roster rows err: %w", err)
	}
	return roster, nil
}

func (r *postgresAnalyticsRepository) topLinks(ctx context.Context, query string, args ...any) ([]models.LinkClickCount, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("analytics top links query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	links := []models.LinkClickCount{}
	for rows.Next() {
		var (
			link                models.LinkClickCount
			linkID, extraLinkID sql.NullInt64
		)
		if err := rows.Scan(&linkID, &extraLinkID, &link.Clicks); err != nil {
			return nil, fmt.Errorf("analytics top links rows scan: %w", err)
		}
		if linkID.Valid {
			id := int(linkID.Int64)
			link.LinkID = &id
		}
		if extraLinkID.Valid {
			id := int(extraLinkID.Int64)
			link.ExtraLinkID = &id
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics top links rows err: %w", err)
	}
	return links, nil
}

func (r *postgresAnalyticsRepository) userClicks(ctx context.Context, query string, args ...any) ([]models.UserClickCount, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	users := []models.UserClickCount{}
	for rows.Next() {
		var (
			user                models.UserClickCount
			firstName, lastName string
			number              int
		)
		if err := rows.Scan(&user.UserID, &firstName, &lastName, &number, &user.Clicks); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		user.Handle = models.UserHandle(firstName, lastName, number, user.UserID)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return users, nil
}

func (r *postgresAnalyticsRepository) visitorsToday(ctx context.Context, params AnalyticsSummaryParams) (models.VisitorsTodayPage, error) {
	query := analyticsVisitorsTodayByClicksQuery
	if params.VisitorsSort == "name" {
		query = analyticsVisitorsTodayByNameQuery
	}

	limit := params.VisitorsLimit
	if limit <= 0 {
		limit = 12
	}
	fetchLimit := limit + 1

	rows, err := r.db.QueryContext(ctx, query, fetchLimit, params.VisitorsOffset)
	if err != nil {
		return models.VisitorsTodayPage{}, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	visitors, err := r.scanUserClickRows(rows)
	if err != nil {
		return models.VisitorsTodayPage{}, err
	}

	hasMore := len(visitors) > limit
	if hasMore {
		visitors = visitors[:limit]
	}
	return models.VisitorsTodayPage{Visitors: visitors, HasMore: hasMore}, nil
}

func (r *postgresAnalyticsRepository) scanUserClickRows(rows *sql.Rows) ([]models.UserClickCount, error) {
	users := []models.UserClickCount{}
	for rows.Next() {
		var (
			user                models.UserClickCount
			firstName, lastName string
			number              int
		)
		if err := rows.Scan(&user.UserID, &firstName, &lastName, &number, &user.Clicks); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		user.Handle = models.UserHandle(firstName, lastName, number, user.UserID)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return users, nil
}

func (r *postgresAnalyticsRepository) courseDemand(ctx context.Context, query string, args ...any) ([]models.CourseDemand, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []models.CourseDemand{}
	for rows.Next() {
		var c models.CourseDemand
		if err := rows.Scan(&c.CourseID, &c.Name, &c.Code, &c.Count, &c.ProgramName); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (r *postgresAnalyticsRepository) deadLinks(ctx context.Context, days int) ([]models.DeadLink, error) {
	rows, err := r.db.QueryContext(ctx, analyticsZeroClickLinksQuery, days)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []models.DeadLink{}
	for rows.Next() {
		var d models.DeadLink
		if err := rows.Scan(&d.Kind, &d.ID, &d.Label, &d.CourseName, &d.ProgramName); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (r *postgresAnalyticsRepository) heatmap(ctx context.Context, days int) ([]models.HeatmapCell, error) {
	rows, err := r.db.QueryContext(ctx, analyticsHeatmapQuery, days)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []models.HeatmapCell{}
	for rows.Next() {
		var c models.HeatmapCell
		if err := rows.Scan(&c.Dow, &c.Hour, &c.Count); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (r *postgresAnalyticsRepository) searchTerms(ctx context.Context, days int) ([]models.SearchTermCount, error) {
	rows, err := r.db.QueryContext(ctx, analyticsSearchTermsQuery, days)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []models.SearchTermCount{}
	for rows.Next() {
		var s models.SearchTermCount
		if err := rows.Scan(&s.Query, &s.Count); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
