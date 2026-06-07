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

type LinkService struct {
	repo repository.LinkRepository
}

func NewLinkService(repo repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) Create(ctx context.Context, link models.Link) error {
	link.URL = strings.TrimSpace(link.URL)
	link.Label = strings.TrimSpace(link.Label)
	if link.Label == "" || link.URL == "" {
		return errs.ErrLinkURLAndLabelRequired
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return fmt.Errorf("create link: %w", err)
	}
	return nil
}

func (s *LinkService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrLinkInvalidID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}

func (s *LinkService) Update(ctx context.Context, link models.Link, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrLinkInvalidID
	}
	if err := s.repo.Update(ctx, link, id); err != nil {
		return fmt.Errorf("update link: %w", err)
	}
	return nil
}
