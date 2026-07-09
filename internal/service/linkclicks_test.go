package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeLinkClickRepo struct {
	createCalls int
	createLC    models.LinkClick
	createErr   error

	listCalls  int
	listResult []models.LinkClick
	listErr    error
}

func (f *fakeLinkClickRepo) Create(ctx context.Context, lc models.LinkClick) error {
	f.createCalls++
	f.createLC = lc
	return f.createErr
}

func (f *fakeLinkClickRepo) List(ctx context.Context) ([]models.LinkClick, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func TestLinkClickService_Create(t *testing.T) {
	tests := []struct {
		name         string
		lc           models.LinkClick
		createErr    error
		err          error
		resultWanted *models.LinkClick
	}{
		{
			name:         "persists link click",
			lc:           models.LinkClick{LinkID: &[]int{42}[0]},
			resultWanted: &models.LinkClick{LinkID: &[]int{42}[0]},
		},
		{
			name:         "persists extra link click",
			lc:           models.LinkClick{ExtraLinkID: &[]int{42}[0]},
			resultWanted: &models.LinkClick{ExtraLinkID: &[]int{42}[0]},
		},
		{
			name: "persists both link and extra link clicks",
			lc:   models.LinkClick{LinkID: &[]int{42}[0], ExtraLinkID: &[]int{42}[0]},
			err:  errs.ErrLinkClickLinkIDAndExtraLinkIDSet,
		},
		{
			name: "link id and extra link id are required",
			lc:   models.LinkClick{LinkID: nil, ExtraLinkID: nil},
			err:  errs.ErrLinkClickLinkIDAndExtraLinkIDRequired,
		},
		{
			name:      "repo create error",
			lc:        models.LinkClick{LinkID: &[]int{42}[0]},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLinkClickRepo{createErr: tt.createErr}
			s := NewLinkClickService(repo)
			err := s.Create(context.Background(), tt.lc)
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
				t.Fatalf("Create Link Click Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createLC, *tt.resultWanted) {
				t.Fatalf("repo.Create link click: got %+v want %+v", repo.createLC, *tt.resultWanted)
			}
		})
	}
}

func TestLinkClickService_List(t *testing.T) {
	sample := []models.LinkClick{
		{ID: 1, LinkID: &[]int{42}[0], ExtraLinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
		{ID: 2, LinkID: &[]int{99}[0], ExtraLinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
		{ID: 3, ExtraLinkID: &[]int{42}[0], LinkID: nil, ClickedAt: "2024-01-01T00:00:00Z"},
		{ID: 4, ExtraLinkID: &[]int{99}[0], LinkID: nil, ClickedAt: "2024-02-01T00:00:00Z"},
	}

	tests := []struct {
		name       string
		repoResult []models.LinkClick
		repoErr    error
		wantResult []models.LinkClick
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
			repo := &fakeLinkClickRepo{listResult: tt.repoResult, listErr: tt.repoErr}
			s := NewLinkClickService(repo)

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
