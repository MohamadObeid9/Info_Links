package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeExtraSectionRepo struct {
	listCalls   int
	listResult  []models.ExtraSection
	listErr     error
	createCalls int
	create      models.ExtraSection
	createErr   error
	updateCalls int
	update      models.ExtraSection
	updateID    int
	updateErr   error
	deleteCalls int
	deleteID    int
	deleteErr   error
}

func (f *fakeExtraSectionRepo) List(ctx context.Context) ([]models.ExtraSection, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeExtraSectionRepo) Create(ctx context.Context, section models.ExtraSection) error {
	f.createCalls++
	f.create = section
	return f.createErr
}

func (f *fakeExtraSectionRepo) Update(ctx context.Context, section models.ExtraSection, id int) error {
	f.updateCalls++
	f.update = section
	f.updateID = id
	return f.updateErr
}

func (f *fakeExtraSectionRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func TestExtraSectionService_Create(t *testing.T) {
	tests := []struct {
		name         string
		section      models.ExtraSection
		createErr    error
		err          error
		resultWanted *models.ExtraSection
	}{
		{
			name:    "title is required",
			section: models.ExtraSection{Title: "", Icon: "📁"},
			err:     errs.ErrExtraSectionTitleRequired,
		},
		{
			name:    "title shouldn't be empty",
			section: models.ExtraSection{Title: "  ", Icon: "📁"},
			err:     errs.ErrExtraSectionTitleRequired,
		},
		{
			name:      "repo create error",
			section:   models.ExtraSection{Title: "Python", Icon: "🐍", DisplayOrder: 0},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "trims and defaults icon",
			section:      models.ExtraSection{Title: "  Python  ", Icon: "  ", DisplayOrder: 0},
			resultWanted: &models.ExtraSection{Title: "Python", Icon: "📁", DisplayOrder: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraSectionRepo{createErr: tt.createErr}
			s := NewExtraSectionService(repo)
			err := s.Create(context.Background(), tt.section)
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
				t.Fatalf("Create succeeded, want error %v", tt.err)
			}
			if !reflect.DeepEqual(repo.create, *tt.resultWanted) {
				t.Fatalf("repo.Create: got %+v want %+v", repo.create, *tt.resultWanted)
			}
		})
	}
}

func TestExtraSectionService_Update(t *testing.T) {
	section := models.ExtraSection{Title: "Updated", Icon: "📁", DisplayOrder: 1}

	tests := []struct {
		name         string
		idStr        string
		id           int
		section      models.ExtraSection
		updateErr    error
		err          error
		resultWanted *models.ExtraSection
	}{
		{
			name:    "reject non numerical id",
			idStr:   "abc",
			section: section,
			err:     errs.ErrExtraSectionInvalidID,
		},
		{
			name:    "reject id = 0",
			idStr:   "0",
			section: section,
			err:     errs.ErrExtraSectionInvalidID,
		},
		{
			name:    "title is required",
			idStr:   "10",
			section: models.ExtraSection{Title: "  ", Icon: "📁"},
			err:     errs.ErrExtraSectionTitleRequired,
		},
		{
			name:      "repo update error",
			idStr:     "10",
			id:        10,
			section:   section,
			updateErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "trims and defaults icon",
			idStr:        "10",
			id:           10,
			section:      models.ExtraSection{Title: "  Updated  ", Icon: "  ", DisplayOrder: 1},
			resultWanted: &section,
		},
		{
			name:         "accept a valid id with spaces",
			idStr:        "  10  ",
			id:           10,
			section:      section,
			resultWanted: &section,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraSectionRepo{updateErr: tt.updateErr}
			s := NewExtraSectionService(repo)
			err := s.Update(context.Background(), tt.section, tt.idStr)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.updateErr != nil {
					wantCalls = 1
				}
				if repo.updateCalls != wantCalls {
					t.Fatalf("repo.Update calls: got %d want %d", repo.updateCalls, wantCalls)
				}
				return
			}
			if repo.updateID != tt.id {
				t.Fatalf("repo.UpdateID: got %d want %d", repo.updateID, tt.id)
			}
			if !reflect.DeepEqual(repo.update, *tt.resultWanted) {
				t.Fatalf("repo.Update: got %+v want %+v", repo.update, *tt.resultWanted)
			}
		})
	}
}

func TestExtraSectionService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		idStr     string
		id        int
		err       error
		deleteErr error
	}{
		{name: "reject non numerical id", idStr: "abc", err: errs.ErrExtraSectionInvalidID},
		{name: "reject empty id with spaces", idStr: "  ", err: errs.ErrExtraSectionInvalidID},
		{name: "reject empty id without spaces", idStr: "", err: errs.ErrExtraSectionInvalidID},
		{name: "reject id = 0", idStr: "0", err: errs.ErrExtraSectionInvalidID},
		{name: "reject id < 0", idStr: "-10", err: errs.ErrExtraSectionInvalidID},
		{name: "repo delete error", idStr: "10", deleteErr: errs.ErrDatabaseDown, err: errs.ErrDatabaseDown},
		{name: "accept a valid id", idStr: "10", id: 10},
		{name: "accept a valid id with spaces", idStr: "  10  ", id: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraSectionRepo{deleteErr: tt.deleteErr}
			s := NewExtraSectionService(repo)
			err := s.Delete(context.Background(), tt.idStr)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				wantCalls := 0
				if tt.deleteErr != nil {
					wantCalls = 1
				}
				if repo.deleteCalls != wantCalls {
					t.Fatalf("repo.Delete calls: got %d want %d", repo.deleteCalls, wantCalls)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Delete ExtraSectionService succeeded, want error %v", tt.err)
			}
			if repo.deleteCalls != 1 {
				t.Fatalf("repo.Delete calls: got %d want 1", repo.deleteCalls)
			}
			if repo.deleteID != tt.id {
				t.Fatalf("repo.DeleteID: got %d want %d", repo.deleteID, tt.id)
			}
		})
	}
}
