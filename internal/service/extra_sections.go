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

type ExtraSectionService struct {
	repo repository.ExtraSectionRepository
}

func NewExtraSectionService(repo repository.ExtraSectionRepository) *ExtraSectionService {
	return &ExtraSectionService{repo: repo}
}

func (s *ExtraSectionService) List(ctx context.Context) ([]models.ExtraSection, error) {
	sections, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list extra sections: %w", err)
	}
	return sections, nil
}

func (s *ExtraSectionService) Create(ctx context.Context, section models.ExtraSection) error {
	section.Title = strings.TrimSpace(section.Title)
	section.Icon = strings.TrimSpace(section.Icon)
	if section.Title == "" {
		return errs.ErrExtraSectionTitleRequired
	}
	if section.Icon == "" {
		section.Icon = "📁"
	}

	if err := s.repo.Create(ctx, section); err != nil {
		return fmt.Errorf("create extra section: %w", err)
	}
	return nil
}

func (s *ExtraSectionService) Update(ctx context.Context, section models.ExtraSection, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrExtraSectionInvalidID
	}

	section.Title = strings.TrimSpace(section.Title)
	section.Icon = strings.TrimSpace(section.Icon)
	if section.Title == "" {
		return errs.ErrExtraSectionTitleRequired
	}
	if section.Icon == "" {
		section.Icon = "📁"
	}

	if err := s.repo.Update(ctx, section, id); err != nil {
		return fmt.Errorf("update extra section: %w", err)
	}
	return nil
}

func (s *ExtraSectionService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrExtraSectionInvalidID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete extra section: %w", err)
	}
	return nil
}
