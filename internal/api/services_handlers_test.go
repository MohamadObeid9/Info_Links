package api

import (
	"context"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeServiceService struct {
	services []models.Service
	err      error
	created  *models.Service
	updated  *models.ServicePatch
	deleted  string
	renewed  string
	frozen   string
	unfrozen string
	clicked  *models.ServiceClick
}

func (f *fakeServiceService) List(ctx context.Context, limit int, offset int, q string) ([]models.Service, error) {
	return f.services, f.err
}

func (f *fakeServiceService) Get(ctx context.Context, idStr string) (models.Service, error) {
	if f.err != nil {
		return models.Service{}, f.err
	}
	for _, s := range f.services {
		if fmt.Sprintf("%d", s.ID) == idStr {
			return s, nil
		}
	}
	return models.Service{}, errs.ErrServiceNotFound
}

func (f *fakeServiceService) Create(ctx context.Context, svc models.Service) error {
	f.created = &svc
	return f.err
}

func (f *fakeServiceService) Update(ctx context.Context, patch models.ServicePatch, idStr string) error {
	f.updated = &patch
	return f.err
}

func (f *fakeServiceService) Delete(ctx context.Context, idStr string) error {
	f.deleted = idStr
	return f.err
}

func (f *fakeServiceService) Renew(ctx context.Context, idStr string, durationDays int) error {
	f.renewed = idStr
	return f.err
}

func (f *fakeServiceService) Freeze(ctx context.Context, idStr string) error {
	f.frozen = idStr
	return f.err
}

func (f *fakeServiceService) Unfreeze(ctx context.Context, idStr string) error {
	f.unfrozen = idStr
	return f.err
}

func (f *fakeServiceService) TrackClick(ctx context.Context, click models.ServiceClick) error {
	f.clicked = &click
	return f.err
}

func (f *fakeServiceService) FreezeExpired(ctx context.Context) error {
	return f.err
}
