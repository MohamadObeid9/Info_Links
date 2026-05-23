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

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) Create(ctx context.Context, course models.Course) error {
	course.Name = strings.TrimSpace(course.Name)
	course.Code = strings.TrimSpace(course.Code)
	if course.Name == "" || course.Code == "" {
		return errs.ErrCourseCodeAndNameRequired
	}

	if err := s.repo.Create(ctx, course); err != nil {
		return fmt.Errorf("create course: %w", err)
	}
	return nil
}

func (s *CourseService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrCourseInvalidID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}

func (s *CourseService) Update(ctx context.Context, course models.Course, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	course.Name = strings.TrimSpace(course.Name)
	course.Code = strings.TrimSpace(course.Code)

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrCourseInvalidID
	}
	if err := s.repo.Update(ctx, course, id); err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}
