package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

const defaultAnalyticsRangeDays = 7

const defaultVisitorsLimit = 12

const maxSearchQueryLen = 80

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

func (s *AnalyticsService) TrackSearch(ctx context.Context, userID int, query string) error {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return errs.ErrAnalyticsInvalidSearchQuery
	}
	if utf8.RuneCountInString(q) > maxSearchQueryLen {
		q = string([]rune(q)[:maxSearchQueryLen])
	}
	if err := s.repo.InsertSearch(ctx, userID, q); err != nil {
		return fmt.Errorf("track search: %w", err)
	}
	return nil
}

func (s *AnalyticsService) ListActors(ctx context.Context, kind, idStr, rangeStr string) (models.AnalyticsActorsResult, error) {
	kind = strings.TrimSpace(kind)
	key := strings.TrimSpace(idStr)

	switch kind {
	case "search":
		key = strings.ToLower(key)
		if key == "" {
			return models.AnalyticsActorsResult{}, errs.ErrAnalyticsInvalidActorID
		}
		if utf8.RuneCountInString(key) > maxSearchQueryLen {
			key = string([]rune(key)[:maxSearchQueryLen])
		}
	case "device":
		if key != "phone" && key != "laptop" && key != "both" {
			return models.AnalyticsActorsResult{}, errs.ErrAnalyticsInvalidActorID
		}
	default:
		id, err := strconv.Atoi(key)
		if err != nil || id <= 0 {
			return models.AnalyticsActorsResult{}, errs.ErrAnalyticsInvalidActorID
		}
		key = strconv.Itoa(id)
	}

	since, err := parseActorsSince(kind, rangeStr)
	if err != nil {
		return models.AnalyticsActorsResult{}, err
	}

	result, err := s.repo.ListActors(ctx, kind, key, since)
	if err != nil {
		return models.AnalyticsActorsResult{}, fmt.Errorf("list actors: %w", err)
	}
	return result, nil
}

func parseActorsSince(kind, rangeStr string) (time.Time, error) {
	if kind == "favorite" {
		return time.Time{}, nil
	}
	switch strings.TrimSpace(rangeStr) {
	case "", "today":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "7", "30", "90":
		days, _ := strconv.Atoi(strings.TrimSpace(rangeStr))
		return time.Now().AddDate(0, 0, -days), nil
	case "all":
		return time.Time{}, nil // zero time = no lower bound in actor queries
	default:
		return time.Time{}, errs.ErrAnalyticsInvalidRange
	}
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
	case "all":
		return 0, nil // Days==0 sentinel: all-time (no day cutoff in SQL)
	default:
		return 0, errs.ErrAnalyticsInvalidRange
	}
}
