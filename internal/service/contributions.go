package service

import (
	"context"
	"fmt"
	"strings"

	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

type ContributionService struct {
	repo repository.ContributionRepository
}

func NewContributionService(repo repository.ContributionRepository) *ContributionService {
	return &ContributionService{repo: repo}
}

func (c *ContributionService) Create(ctx context.Context, contribution models.Contribution) error {
	contribution.CourseName = strings.TrimSpace(contribution.CourseName)
	contribution.LinkURL = strings.TrimSpace(contribution.LinkURL)
	contribution.Note = strings.TrimSpace(contribution.Note)
	if contribution.CourseName == "" || contribution.LinkURL == "" {
		return fmt.Errorf("course_name and link_url are required")
	}

	if err := c.repo.Create(ctx, contribution); err != nil {
		return fmt.Errorf("create contribution: %w", err)
	}
	return nil
}
