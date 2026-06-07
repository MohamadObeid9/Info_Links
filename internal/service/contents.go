package service

import (
	"context"
	"fmt"

	"infolinks-backend/internal/repository"
)

type ContentService struct {
	repo repository.ContentRepository
}

func NewContentService(repo repository.ContentRepository) *ContentService {
	return &ContentService{repo: repo}
}

func (c *ContentService) Get(ctx context.Context) ([]byte, error) {
	result, err := c.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get content: %w", err)
	}
	return result, nil
}
