package service

import (
	"context"
	"fmt"

	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

type PageViewService struct {
	repo repository.PageViewRepository
}

func NewPageViewService(repo repository.PageViewRepository) *PageViewService {
	return &PageViewService{repo: repo}
}

func (s *PageViewService) Create(ctx context.Context, pv models.PageView) error {
	if err := s.repo.Create(ctx, pv); err != nil {
		return fmt.Errorf("create page view: %w", err)
	}
	return nil
}

func (s *PageViewService) List(ctx context.Context) ([]models.PageView, error) {
	views, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list page views: %w", err)
	}
	return views, nil
}
