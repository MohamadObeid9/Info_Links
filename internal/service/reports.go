package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

type ReportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Create(ctx context.Context, report models.Report) error {
	report.CourseName = strings.TrimSpace(report.CourseName)
	report.LinkURL = strings.TrimSpace(report.LinkURL)
	report.Description = strings.TrimSpace(report.Description)
	if report.CourseName == "" || report.LinkURL == "" {
		return errs.ErrCourseNameAndLinkUrlAreRequired
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	return nil
}

func (s *ReportService) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, errs.ErrListReportInvalidParams
	}

	switch status {
	case "open", "resolved":
	case "":
		return nil, errs.ErrReportStatusRequired
	default:
		return nil, errs.ErrInvalidReportStatus
	}

	reps, err := s.repo.List(ctx, limit, offset, q, status)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}

	return reps, nil
}

func (s *ReportService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrInvalidReportID // err 400
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete report: %w", err) // err 500
	}
	return nil
}

func (s *ReportService) Update(ctx context.Context, status string, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	status = strings.TrimSpace(status)

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrInvalidReportID
	}

	switch status {
	case "open", "resolved":
	case "":
		return errs.ErrReportStatusRequired
	default:
		return errs.ErrInvalidReportStatus
	}

	if err := s.repo.Update(ctx, status, id); err != nil {
		return fmt.Errorf("update report: %w", err)
	}
	return nil
}
