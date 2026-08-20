package service

import (
	"context"
	"fmt"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

const defaultAnalyticsRangeDays = 7

const defaultVisitorsLimit = 12

type AnalyticsVisitorsParams struct {
	Limit  int
	Offset int
	Sort   string
}

type AnalyticsService struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetSummary(ctx context.Context, rangeStr string, visitors AnalyticsVisitorsParams) (models.AnalyticsSummary, error) {
	days, err := parseAnalyticsRange(rangeStr)
	if err != nil {
		return models.AnalyticsSummary{}, err
	}

	sort := strings.TrimSpace(visitors.Sort)
	if sort == "" {
		sort = "clicks"
	}
	if sort != "clicks" && sort != "name" {
		return models.AnalyticsSummary{}, errs.ErrAnalyticsInvalidVisitorsSort
	}

	limit := visitors.Limit
	if limit <= 0 {
		limit = defaultVisitorsLimit
	}
	if limit > 100 {
		limit = 100
	}
	offset := visitors.Offset
	if offset < 0 {
		offset = 0
	}

	summary, err := s.repo.GetSummary(ctx, repository.AnalyticsSummaryParams{
		Days:           days,
		VisitorsLimit:  limit,
		VisitorsOffset: offset,
		VisitorsSort:   sort,
	})
	if err != nil {
		return models.AnalyticsSummary{}, fmt.Errorf("get analytics summary: %w", err)
	}
	return summary, nil
}

func parseAnalyticsRange(rangeStr string) (int, error) {
	switch strings.TrimSpace(rangeStr) {
	case "":
		return defaultAnalyticsRangeDays, nil
	case "7":
		return 7, nil
	case "30":
		return 30, nil
	case "90":
		return 90, nil
	default:
		return 0, errs.ErrAnalyticsInvalidRange
	}
}
