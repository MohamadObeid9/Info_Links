package service

import (
	"context"
	"fmt"
	"strings"

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
		return fmt.Errorf("course_name and link_url are required")
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	return nil
}
