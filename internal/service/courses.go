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
	if course.SemesterID <= 0 {
		return errs.ErrCourseInvalidSemestreID
	}

	if err := s.repo.Create(ctx, course); err != nil {
		return fmt.Errorf("create course: %w", err)
	}
	return nil
}

func (s *CourseService) Delete(ctx context.Context, idStr, placementStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrCourseInvalidID
	}
	placementStr = strings.TrimSpace(placementStr)
	if placementStr == "" {
		if err := s.repo.Delete(ctx, id); err != nil {
			return fmt.Errorf("delete course: %w", err)
		}
		return nil
	}
	placementID, err := strconv.Atoi(placementStr)
	if err != nil || placementID <= 0 {
		return errs.ErrCourseInvalidPlacementID
	}
	if err := s.repo.DeletePlacement(ctx, id, placementID); err != nil {
		return fmt.Errorf("delete course placement: %w", err)
	}
	return nil
}

func (s *CourseService) Update(ctx context.Context, patch models.CoursePatch, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrCourseInvalidID
	}
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}

	merged := existing

	if patch.Name == nil && patch.Code == nil && patch.SemesterID == nil && patch.IsOptional == nil && patch.PlacementID == nil {
		return errs.ErrCoursePatchEmpty
	}

	if patch.Name != nil {
		merged.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Code != nil {
		merged.Code = strings.TrimSpace(*patch.Code)
	}
	if patch.PlacementID != nil {
		if *patch.PlacementID <= 0 {
			return errs.ErrCourseInvalidPlacementID
		}
		merged.PlacementID = *patch.PlacementID
	}
	if patch.SemesterID != nil {
		if *patch.SemesterID <= 0 {
			return errs.ErrCourseInvalidSemestreID
		}
		merged.SemesterID = *patch.SemesterID
		if merged.PlacementID <= 0 {
			return errs.ErrCourseInvalidPlacementID
		}
	}
	if patch.IsOptional != nil {
		merged.IsOptional = *patch.IsOptional
	}

	if merged.Name == "" || merged.Code == "" {
		return errs.ErrCourseCodeAndNameRequired
	}
	if merged.PlacementID > 0 && merged.SemesterID <= 0 {
		return errs.ErrCourseInvalidSemestreID
	}

	if err := s.repo.Update(ctx, merged, id); err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}
