package api

import (
	"context"
	"infolinks-backend/internal/models"
)

type fakeLinkClickService struct{}

func (s *fakeLinkClickService) Create(ctx context.Context, lc models.LinkClick) error {
	return nil
}

func (s *fakeLinkClickService) List(ctx context.Context) ([]models.LinkClick, error) {
	return nil, nil
}
