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

type ExtraLinkService struct {
	repo repository.ExtraLinkRepository
}

func NewExtraLinkService(repo repository.ExtraLinkRepository) *ExtraLinkService {
	return &ExtraLinkService{repo: repo}
}

func (s *ExtraLinkService) List(ctx context.Context) ([]models.ExtraLink, error) {
	links, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list extra links: %w", err)
	}
	return links, nil
}

func (s *ExtraLinkService) Create(ctx context.Context, link models.ExtraLink) error {
	link.URL = strings.TrimSpace(link.URL)
	link.Label = strings.TrimSpace(link.Label)
	if link.Label == "" || link.URL == "" {
		return errs.ErrExtraLinkURLAndLabelRequired
	}
	if link.SectionID == nil || *link.SectionID <= 0 {
		return errs.ErrExtraLinkInvalidSectionID
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return fmt.Errorf("create extra link: %w", err)
	}
	return nil
}

func (s *ExtraLinkService) Update(ctx context.Context, link models.ExtraLink, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrExtraLinkInvalidID
	}
	if err := s.repo.Update(ctx, link, id); err != nil {
		return fmt.Errorf("update extra link: %w", err)
	}
	return nil
}

func (s *ExtraLinkService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrExtraLinkInvalidID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete extra link: %w", err)
	}
	return nil
}
