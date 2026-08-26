package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/service"
	"infolinks-backend/internal/webbotauth"
)

type Handler struct {
	logger              *slog.Logger
	jwtSecret           []byte
	siteBaseURL         string
	supabaseURL         string
	supbaseAnonKey      string
	httpClient          *http.Client
	webBotAuth          *webbotauth.Directory
	db                  dbPinger
	linkService         linkService
	userService         userService
	analyticsService    analyticsService
	reportService       reportService
	courseService       courseService
	contentService      contentService
	pageViewService     pageViewService
	feedbackService     feedbackService
	linkClickService    linkClickService
	contributionService contributionService
	extraSectionService extraSectionService
	extraLinkService    extraLinkService
}

type Dependencies struct {
	Logger              *slog.Logger
	JWTSecret           []byte
	SiteBaseURL         string
	SupabaseURL         string
	SupabaseAnonKey     string
	WebBotAuth          *webbotauth.Directory
	DB                  dbPinger
	LinkService         linkService
	UserService         userService
	AnalyticsService    analyticsService
	ReportService       reportService
	CourseService       courseService
	ContentService      contentService
	FeedbackService     feedbackService
	PageViewService     pageViewService
	LinkClickService    linkClickService
	ContributionService contributionService
	ExtraSectionService extraSectionService
	ExtraLinkService    extraLinkService
}

type dbPinger interface {
	Ping(ctx context.Context) error
}

type contentService interface {
	Get(ctx context.Context) ([]byte, error)
}

type userService interface {
	CreateGuest(ctx context.Context) (int, error)
	RegisterUser(ctx context.Context, guestID int, u models.User) (models.User, error)
	LoginUser(ctx context.Context, guestID int, u models.User) (models.User, error)
	GetUser(ctx context.Context, userID int) (models.User, error)
	AddFavorite(ctx context.Context, userID int, courseIDStr string) error
	RemoveFavorite(ctx context.Context, userID int, courseIDStr string) error
	ListStudents(ctx context.Context, limit int, offset int, q string) ([]models.UserListItem, error)
	GetUserDetail(ctx context.Context, idStr string, limit int, offset int) (models.UserDetail, error)
}

type analyticsService interface {
	GetSummary(ctx context.Context, rangeStr string, visitors service.AnalyticsVisitorsParams) (models.AnalyticsSummary, error)
	TrackSearch(ctx context.Context, userID int, query string) error
	TrackBrowse(ctx context.Context, userID int, step string) error
}

type pageViewService interface {
	List(ctx context.Context) ([]models.PageView, error)
	Create(ctx context.Context, pv models.PageView) error
}

type linkClickService interface {
	List(ctx context.Context) ([]models.LinkClick, error)
	Create(ctx context.Context, lc models.LinkClick) error
}

type linkService interface {
	Delete(ctx context.Context, idStr string) error
	Create(ctx context.Context, link models.Link) error
	Update(ctx context.Context, link models.Link, idStr string) error
}

type courseService interface {
	Delete(ctx context.Context, idStr, placementStr string) error
	Create(ctx context.Context, course models.Course) error
	Update(ctx context.Context, patch models.CoursePatch, idStr string) error
}

type reportService interface {
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, report models.Report) error
	Update(ctx context.Context, status string, idStr string) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error)
}

type feedbackService interface {
	Delete(ctx context.Context, idStr string) error
	Create(ctx context.Context, feedback models.Feedback) error
	Update(ctx context.Context, status string, idStr string) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error)
}

type contributionService interface {
	Delete(ctx context.Context, idStr string) error
	Update(ctx context.Context, status string, idStr string) error
	Create(ctx context.Context, contribution models.Contribution) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error)
}

type extraSectionService interface {
	List(ctx context.Context) ([]models.ExtraSection, error)
	Create(ctx context.Context, section models.ExtraSection) error
	Update(ctx context.Context, section models.ExtraSection, idStr string) error
	Delete(ctx context.Context, idStr string) error
}

type extraLinkService interface {
	List(ctx context.Context) ([]models.ExtraLink, error)
	Create(ctx context.Context, link models.ExtraLink) error
	Update(ctx context.Context, link models.ExtraLink, idStr string) error
	Delete(ctx context.Context, idStr string) error
}

func NewHandler(deps Dependencies) (*Handler, error) {

	if deps.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if len(deps.JWTSecret) == 0 || strings.TrimSpace(string(deps.JWTSecret)) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	if deps.SupabaseURL == "" {
		return nil, fmt.Errorf("supabase url is required")
	}

	if deps.SupabaseAnonKey == "" {
		return nil, fmt.Errorf("supabase anon key is required")
	}

	if deps.DB == nil {
		return nil, fmt.Errorf("db pinger is required")
	}

	if deps.ReportService == nil {
		return nil, fmt.Errorf("report service is required")
	}

	if deps.UserService == nil {
		return nil, fmt.Errorf("user service is required")
	}

	if deps.AnalyticsService == nil {
		return nil, fmt.Errorf("analytics service is required")
	}

	if deps.ContributionService == nil {
		return nil, fmt.Errorf("contribution service is required")
	}

	if deps.ContentService == nil {
		return nil, fmt.Errorf("content service is required")
	}

	if deps.FeedbackService == nil {
		return nil, fmt.Errorf("feedback service is required ")
	}

	if deps.CourseService == nil {
		return nil, fmt.Errorf("course service is required")
	}

	if deps.LinkClickService == nil {
		return nil, fmt.Errorf("link click servie is required")
	}

	if deps.LinkService == nil {
		return nil, fmt.Errorf("link service is required")
	}

	if deps.ExtraSectionService == nil {
		return nil, fmt.Errorf("extra section service is required")
	}

	if deps.ExtraLinkService == nil {
		return nil, fmt.Errorf("extra link service is required")
	}

	newHandler := Handler{
		db:                  deps.DB,
		logger:              deps.Logger,
		jwtSecret:           deps.JWTSecret,
		siteBaseURL:         strings.TrimSuffix(strings.TrimSpace(deps.SiteBaseURL), "/"),
		webBotAuth:          deps.WebBotAuth,
		linkService:         deps.LinkService,
		supabaseURL:         deps.SupabaseURL,
		userService:         deps.UserService,
		analyticsService:    deps.AnalyticsService,
		courseService:       deps.CourseService,
		reportService:       deps.ReportService,
		contentService:      deps.ContentService,
		feedbackService:     deps.FeedbackService,
		pageViewService:     deps.PageViewService,
		supbaseAnonKey:      deps.SupabaseAnonKey,
		linkClickService:    deps.LinkClickService,
		extraLinkService:    deps.ExtraLinkService,
		contributionService: deps.ContributionService,
		extraSectionService: deps.ExtraSectionService,
		httpClient:          http.DefaultClient,
	}

	return &newHandler, nil
}

// skipForAdmin answers 204 to analytics POSTs carrying a valid admin token, so
// admins browsing the site never show up in usage stats. It runs before student
// auth because an admin token has no student claims.
func (h *Handler) skipForAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if middleware.IsAuthenticatedAdmin(string(h.jwtSecret), r.Header.Get("Authorization")) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// LoggerWithID returns a logger enriched with the request ID when present.
func (h *Handler) LoggerWithID(r *http.Request) *slog.Logger {
	if id := middleware.RequestIDFromContext(r.Context()); id != "" {
		return h.logger.With("request_id", id)
	}
	return h.logger
}
