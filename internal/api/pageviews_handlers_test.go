package api

import (
	"context"
	"infolinks-backend/internal/models"
)

type fakePageViewService struct{}

func (s *fakePageViewService) Create(ctx context.Context, pv models.PageView) error {
	return nil
}

func (s *fakePageViewService) List(ctx context.Context) ([]models.PageView, error) {
	return nil, nil
}
