package api

import (
	"context"
	"infolinks-backend/internal/models"
)

type fakeLinkService struct{}

func (s *fakeLinkService) Create(ctx context.Context, link models.Link) error {
	return nil
}

func (s *fakeLinkService) Delete(ctx context.Context, idStr string) error {
	return nil
}

func (s *fakeLinkService) Update(ctx context.Context, link models.Link, idStr string) error {
	return nil
}
