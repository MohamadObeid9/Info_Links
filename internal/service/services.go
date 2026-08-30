package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

const (
	DefaultTrialDays   = 15
	DefaultRenewalDays = 30
)

type ServiceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) List(ctx context.Context, limit int, offset int, q string) ([]models.Service, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, errs.ErrInvalidParams
	}
	if err := s.repo.FreezeExpired(ctx); err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	services, err := s.repo.List(ctx, limit, offset, strings.TrimSpace(q))
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return services, nil
}

func (s *ServiceService) Get(ctx context.Context, idStr string) (models.Service, error) {
	id, err := parseServiceID(idStr)
	if err != nil {
		return models.Service{}, err
	}
	service, err := s.repo.Get(ctx, id)
	if err != nil {
		return models.Service{}, fmt.Errorf("get service: %w", err)
	}
	return service, nil
}

func (s *ServiceService) Create(ctx context.Context, svc models.Service) error {
	svc = normalizeService(svc)
	if svc.Title == "" {
		return errs.ErrServiceTitleRequired
	}
	if svc.Status == "" {
		svc.Status = "trial"
	}
	if !isValidServiceStatus(svc.Status) {
		return errs.ErrServiceInvalidStatus
	}
	applyTrialFromStatus(&svc)
	if svc.StartedAt == "" {
		svc.StartedAt = time.Now().Format(time.RFC3339)
	}
	if svc.ExpiresAt == "" {
		days := DefaultTrialDays
		if !svc.Trial {
			days = DefaultRenewalDays
		}
		svc.ExpiresAt = time.Now().AddDate(0, 0, days).Format(time.RFC3339)
	}
	if _, err := s.repo.Create(ctx, svc); err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	return nil
}

func (s *ServiceService) Update(ctx context.Context, patch models.ServicePatch, idStr string) error {
	id, err := parseServiceID(idStr)
	if err != nil {
		return err
	}
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}

	merged := existing

	if patch.Title == nil && patch.OwnerName == nil && patch.Category == nil && patch.Emoji == nil &&
		patch.Description == nil && patch.LogoURL == nil && patch.Phone == nil && patch.URL == nil &&
		patch.Links == nil && patch.Status == nil && patch.Trial == nil &&
		patch.StartedAt == nil && patch.ExpiresAt == nil && patch.DisplayOrder == nil {
		return errs.ErrServicePatchEmpty
	}

	if patch.Title != nil {
		merged.Title = strings.TrimSpace(*patch.Title)
	}
	if patch.OwnerName != nil {
		merged.OwnerName = strings.TrimSpace(*patch.OwnerName)
	}
	if patch.Category != nil {
		merged.Category = strings.TrimSpace(*patch.Category)
	}
	if patch.Emoji != nil {
		merged.Emoji = strings.TrimSpace(*patch.Emoji)
	}
	if patch.Description != nil {
		merged.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.LogoURL != nil {
		merged.LogoURL = strings.TrimSpace(*patch.LogoURL)
	}
	if patch.Phone != nil {
		merged.Phone = strings.TrimSpace(*patch.Phone)
	}
	if patch.URL != nil {
		merged.URL = strings.TrimSpace(*patch.URL)
	}
	if patch.Links != nil {
		merged.Links = *patch.Links
	}
	if patch.Status != nil {
		status := strings.TrimSpace(*patch.Status)
		if !isValidServiceStatus(status) {
			return errs.ErrServiceInvalidStatus
		}
		merged.Status = status
		applyTrialFromStatus(&merged)
	}
	if patch.StartedAt != nil {
		merged.StartedAt = *patch.StartedAt
	}
	if patch.ExpiresAt != nil {
		merged.ExpiresAt = *patch.ExpiresAt
	}
	if patch.DisplayOrder != nil {
		merged.DisplayOrder = *patch.DisplayOrder
	}

	if merged.Title == "" {
		return errs.ErrServiceTitleRequired
	}
	if err := s.repo.Update(ctx, merged, id); err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	return nil
}

func (s *ServiceService) Delete(ctx context.Context, idStr string) error {
	id, err := parseServiceID(idStr)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

func (s *ServiceService) Renew(ctx context.Context, idStr string, durationDays int) error {
	id, err := parseServiceID(idStr)
	if err != nil {
		return err
	}
	_ = durationDays // renew always extends one month from the previous expiry date
	if err := s.repo.Renew(ctx, id); err != nil {
		return fmt.Errorf("renew service: %w", err)
	}
	return nil
}

func (s *ServiceService) Freeze(ctx context.Context, idStr string) error {
	return s.setStatus(ctx, idStr, "frozen")
}

func (s *ServiceService) Unfreeze(ctx context.Context, idStr string) error {
	return s.setStatus(ctx, idStr, "active")
}

func (s *ServiceService) setStatus(ctx context.Context, idStr, status string) error {
	id, err := parseServiceID(idStr)
	if err != nil {
		return err
	}
	if !isValidServiceStatus(status) {
		return errs.ErrServiceInvalidStatus
	}
	if err := s.repo.SetStatus(ctx, id, status); err != nil {
		return fmt.Errorf("set service status: %w", err)
	}
	return nil
}

func (s *ServiceService) TrackClick(ctx context.Context, click models.ServiceClick) error {
	if click.ServiceID <= 0 {
		return errs.ErrServiceInvalidID
	}
	NormalizeServiceClick(&click)
	if err := s.repo.InsertClick(ctx, click); err != nil {
		return fmt.Errorf("track service click: %w", err)
	}
	return nil
}

func (s *ServiceService) FreezeExpired(ctx context.Context) error {
	if err := s.repo.FreezeExpired(ctx); err != nil {
		return fmt.Errorf("freeze expired services: %w", err)
	}
	return nil
}

func normalizeService(svc models.Service) models.Service {
	svc.Title = strings.TrimSpace(svc.Title)
	svc.OwnerName = strings.TrimSpace(svc.OwnerName)
	svc.Category = strings.TrimSpace(svc.Category)
	svc.Emoji = strings.TrimSpace(svc.Emoji)
	svc.Description = strings.TrimSpace(svc.Description)
	svc.LogoURL = strings.TrimSpace(svc.LogoURL)
	svc.Phone = strings.TrimSpace(svc.Phone)
	svc.URL = strings.TrimSpace(svc.URL)
	svc.Status = strings.TrimSpace(svc.Status)
	return svc
}

func applyTrialFromStatus(svc *models.Service) {
	switch svc.Status {
	case "trial":
		svc.Trial = true
	case "active":
		svc.Trial = false
	}
}

func isValidServiceStatus(status string) bool {
	return status == "trial" || status == "active" || status == "frozen"
}

func parseServiceID(idStr string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil || id <= 0 {
		return 0, errs.ErrServiceInvalidID
	}
	return id, nil
}
