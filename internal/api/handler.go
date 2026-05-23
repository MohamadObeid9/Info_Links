package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"infolinks-backend/internal/models"
)

type Handler struct {
	logger              *slog.Logger
	jwtSecret           []byte
	reportService       reportService
	contentService      contentService
	contributionService contributionService
	feedbackService     feedbackService
	courseService       courseService
	linkClickService    linkClickService
	pageViewService     pageViewService
}

type Dependencies struct {
	Logger              *slog.Logger
	JWTSecret           []byte
	ReportService       reportService
	ContentService      contentService
	ContributionService contributionService
	FeedbackService     feedbackService
	CourseService       courseService
	LinkClickService    linkClickService
	PageViewService     pageViewService
}

type contentService interface {
	Get(ctx context.Context) ([]byte, error)
}

type pageViewService interface {
	List(ctx context.Context) ([]models.PageView, error)
	Create(ctx context.Context, pv models.PageView) error
}

type linkClickService interface {
	List(ctx context.Context) ([]models.LinkClick, error)
	Create(ctx context.Context, lc models.LinkClick) error
}

type reportService interface {
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, report models.Report) error
	Update(ctx context.Context, status string, idStr string) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error)
}

type contributionService interface {
	Delete(ctx context.Context, idStr string) error
	Update(ctx context.Context, status string, idStr string) error
	Create(ctx context.Context, contribution models.Contribution) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error)
}

type feedbackService interface {
	Delete(ctx context.Context, idStr string) error
	Create(ctx context.Context, feedback models.Feedback) error
	Update(ctx context.Context, status string, idStr string) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error)
}

type courseService interface {
	Delete(ctx context.Context, idStr string) error
	Create(ctx context.Context, course models.Course) error
	Update(ctx context.Context, course models.Course, idStr string) error
}

func NewHandler(deps Dependencies) (*Handler, error) {

	if deps.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if len(deps.JWTSecret) == 0 || strings.TrimSpace(string(deps.JWTSecret)) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	if deps.ReportService == nil {
		return nil, fmt.Errorf("report service is required")
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

	newHandler := Handler{
		logger:              deps.Logger,
		jwtSecret:           deps.JWTSecret,
		courseService:       deps.CourseService,
		reportService:       deps.ReportService,
		contentService:      deps.ContentService,
		feedbackService:     deps.FeedbackService,
		pageViewService:     deps.PageViewService,
		linkClickService:    deps.LinkClickService,
		contributionService: deps.ContributionService,
	}

	return &newHandler, nil
}
