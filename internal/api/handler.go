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
	contributionService contributionService
}

type Dependencies struct {
	Logger              *slog.Logger
	JWTSecret           []byte
	ReportService       reportService
	ContributionService contributionService
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

	newHandler := Handler{
		logger:              deps.Logger,
		jwtSecret:           deps.JWTSecret,
		reportService:       deps.ReportService,
		contributionService: deps.ContributionService,
	}

	return &newHandler, nil
}
