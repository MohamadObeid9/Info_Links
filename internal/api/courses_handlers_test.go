package api

import (
	"context"
	"infolinks-backend/internal/models"
)

type fakeCourseService struct{}

func (s *fakeCourseService) Create(ctx context.Context, course models.Course) error {
	return nil
}

func (s *fakeCourseService) Delete(ctx context.Context, idStr string) error {
	return nil
}

func (s *fakeCourseService) Update(ctx context.Context, course models.Course, idStr string) error {
	return nil
}
