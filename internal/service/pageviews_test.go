package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakePageViewRepo struct {
	createCalls int
	createPV    models.PageView
	createErr   error

	listCalls  int
	listResult []models.PageView
	listErr    error
}

func (f *fakePageViewRepo) Create(ctx context.Context, pv models.PageView) error {
	f.createCalls++
	f.createPV = pv
	return f.createErr
}

func (f *fakePageViewRepo) List(ctx context.Context) ([]models.PageView, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func TestPageViewService_Create(t *testing.T) {
	tests := []struct {
		name         string
		pv           models.PageView
		createErr    error
		err          error
		resultWanted *models.PageView
	}{
		{
			name:         "persists page view",
			pv:           models.PageView{Page: "home"},
			resultWanted: &models.PageView{Page: "home"},
		},
		{
			name:      "repo create error",
			pv:        models.PageView{Page: "home"},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakePageViewRepo{createErr: tt.createErr}
			s := NewPageViewService(repo)
			err := s.Create(context.Background(), tt.pv)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.createErr != nil {
					wantCalls = 1
				}
				if repo.createCalls != wantCalls {
					t.Fatalf("repo.Create calls: got %d want %d", repo.createCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Create Page View Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createPV, *tt.resultWanted) {
				t.Fatalf("repo.Create page view: got %+v want %+v", repo.createPV, *tt.resultWanted)
			}
		})
	}
}

func TestPageViewService_List(t *testing.T) {
	sample := []models.PageView{
		{ID: 1, Page: "home", VisitedAt: "2024-01-01T00:00:00Z"},
		{ID: 2, Page: "admin", VisitedAt: "2024-02-01T00:00:00Z"},
	}

	tests := []struct {
		name       string
		repoResult []models.PageView
		repoErr    error
		wantResult []models.PageView
		wantErr    error
		wantCalls  int
	}{
		{
			name:       "returns repo result",
			repoResult: sample,
			wantResult: sample,
			wantCalls:  1,
		},
		{
			name:       "returns empty list",
			repoResult: nil,
			wantResult: nil,
			wantCalls:  1,
		},
		{
			name:      "wraps repo error",
			repoErr:   errs.ErrDatabaseDown,
			wantErr:   errs.ErrDatabaseDown,
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakePageViewRepo{listResult: tt.repoResult, listErr: tt.repoErr}
			s := NewPageViewService(repo)

			got, err := s.List(context.Background())
			if repo.listCalls != tt.wantCalls {
				t.Fatalf("repo.List calls = %d, want %d", repo.listCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Fatalf("got %+v, want %+v", got, tt.wantResult)
			}
		})
	}
}
