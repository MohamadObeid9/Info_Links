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

type ContributionService struct {
	repo repository.ContributionRepository
}

func NewContributionService(repo repository.ContributionRepository) *ContributionService {
	return &ContributionService{repo: repo}
}

func (c *ContributionService) Create(ctx context.Context, contribution models.Contribution) error {
	contribution.CourseName = strings.TrimSpace(contribution.CourseName)
	contribution.LinkURL = strings.TrimSpace(contribution.LinkURL)
	contribution.LinkType = strings.TrimSpace(contribution.LinkType)
	contribution.Note = strings.TrimSpace(contribution.Note)
	if contribution.CourseName == "" || contribution.LinkURL == "" {
		return errs.ErrCourseNameAndLinkUrlRequired
	}

	if err := c.repo.Create(ctx, contribution); err != nil {
		return fmt.Errorf("create contribution: %w", err)
	}
	return nil
}

func (c *ContributionService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrInvalidContributionID
	}
	if err := c.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete contribution: %w", err)
	}
	return nil
}

func (c *ContributionService) Update(ctx context.Context, status string, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	status = strings.TrimSpace(status)

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrInvalidContributionID
	}

	switch status {
	case "pending", "approved":
	case "":
		return errs.ErrStatusRequired
	default:
		return errs.ErrInvalidContributionStatus
	}

	if err := c.repo.Update(ctx, status, id); err != nil {
		return fmt.Errorf("update contribution: %w", err)
	}
	return nil
}

func (c *ContributionService) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, errs.ErrInvalidParams
	}

	switch status {
	case "pending", "approved", "":
	default:
		return nil, errs.ErrInvalidContributionStatus
	}

	contributions, err := c.repo.List(ctx, limit, offset, q, status)
	if err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}

	return contributions, nil
}
