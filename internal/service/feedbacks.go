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

type FeedbackService struct {
	repo repository.FeedbackRepository
}

func NewFeedbackService(repo repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

func (s *FeedbackService) Create(ctx context.Context, feedback models.Feedback) error {
	feedback.Category = strings.TrimSpace(feedback.Category)
	feedback.Message = strings.TrimSpace(feedback.Message)
	if feedback.Category == "" || feedback.Rating == 0 {
		return errs.ErrFeedbackCategoryAndRatingRequired
	}

	switch feedback.Category {
	case "ui/ux", "content", "functionality", "performance", "accessibility":
	default:
		return errs.ErrFeedbackInvalidCategory
	}

	switch feedback.Rating {
	case 1, 2, 3, 4, 5:
	default:
		return errs.ErrFeedbackInvalidRating
	}

	if err := s.repo.Create(ctx, feedback); err != nil {
		return fmt.Errorf("create feedback: %w", err)
	}
	return nil
}

func (s *FeedbackService) Update(ctx context.Context, status string, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	status = strings.TrimSpace(status)

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrFeedbackInvalidID
	}

	switch status {
	case "new", "read", "rejected":
	case "":
		return errs.ErrStatusRequired
	default:
		return errs.ErrFeedbackInvalidStatus
	}

	if err := s.repo.Update(ctx, status, id); err != nil {
		return fmt.Errorf("update feedback: %w", err)
	}
	return nil
}

func (s *FeedbackService) Delete(ctx context.Context, idStr string) error {
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return errs.ErrFeedbackInvalidID
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	return nil
}

func (s *FeedbackService) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, errs.ErrInvalidParams
	}

	switch status {
	case "new", "read", "rejected", "":
	default:
		return nil, errs.ErrFeedbackInvalidStatus
	}

	feedbacks, err := s.repo.List(ctx, limit, offset, q, status)
	if err != nil {
		return nil, fmt.Errorf("list feedbacks: %w", err)
	}

	return feedbacks, nil
}
