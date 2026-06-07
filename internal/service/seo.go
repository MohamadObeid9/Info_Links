package service

import (
	"context"
	"fmt"

	"infolinks-backend/internal/repository"
)

type SEOService struct {
	repo repository.SEORepository
}

func NewSEOService(repo repository.SEORepository) *SEOService {
	return &SEOService{repo: repo}
}

func (s *SEOService) GetCoursePageByCode(ctx context.Context, code string) (*repository.CoursePageData, error) {
	data, err := s.repo.GetCoursePageByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get seo course page: %w", err)
	}
	return data, nil
}

func (s *SEOService) ListCourseCodesForSitemap(ctx context.Context) ([]string, error) {
	codes, err := s.repo.ListCourseCodesForSitemap(ctx)
	if err != nil {
		return nil, fmt.Errorf("list seo course codes: %w", err)
	}
	return codes, nil
}

func (s *SEOService) ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]repository.ProgramSitemapEntry, error) {
	programs, err := s.repo.ListProgramsForSitemap(ctx, slugFn)
	if err != nil {
		return nil, fmt.Errorf("list seo programs for sitemap: %w", err)
	}
	return programs, nil
}

func (s *SEOService) ListCoursesIndex(ctx context.Context) ([]repository.CourseIndexEntry, error) {
	entries, err := s.repo.ListCoursesIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("list seo courses index: %w", err)
	}
	return entries, nil
}

func (s *SEOService) GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*repository.ProgramPageData, error) {
	data, err := s.repo.GetProgramBySlug(ctx, slug, slugFn)
	if err != nil {
		return nil, fmt.Errorf("get seo program page: %w", err)
	}
	return data, nil
}
