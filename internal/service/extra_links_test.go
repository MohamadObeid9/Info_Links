package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeExtraLinkRepo struct {
	listCalls   int
	listResult  []models.ExtraLink
	listErr     error
	createCalls int
	createLink  models.ExtraLink
	createErr   error
	updateCalls int
	updateLink  models.ExtraLink
	updateID    int
	updateErr   error
	deleteCalls int
	deleteID    int
	deleteErr   error
}

func (f *fakeExtraLinkRepo) List(ctx context.Context) ([]models.ExtraLink, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeExtraLinkRepo) Create(ctx context.Context, link models.ExtraLink) error {
	f.createCalls++
	f.createLink = link
	return f.createErr
}

func (f *fakeExtraLinkRepo) Update(ctx context.Context, link models.ExtraLink, id int) error {
	f.updateCalls++
	f.updateLink = link
	f.updateID = id
	return f.updateErr
}

func (f *fakeExtraLinkRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func extraSectionID(v int) *int { return &v }

func TestExtraLinkService_Create(t *testing.T) {
	tests := []struct {
		name         string
		link         models.ExtraLink
		createErr    error
		err          error
		resultWanted *models.ExtraLink
	}{
		{
			name: "label is required",
			link: models.ExtraLink{SectionID: extraSectionID(1), URL: "https://fake.test", Label: ""},
			err:  errs.ErrExtraLinkURLAndLabelRequired,
		},
		{
			name: "url is required",
			link: models.ExtraLink{SectionID: extraSectionID(1), URL: "", Label: "link 1"},
			err:  errs.ErrExtraLinkURLAndLabelRequired,
		},
		{
			name: "section id is required",
			link: models.ExtraLink{URL: "https://fake.test", Label: "link 1"},
			err:  errs.ErrExtraLinkInvalidSectionID,
		},
		{
			name:      "repo create error",
			link:      models.ExtraLink{SectionID: extraSectionID(1), Type: "telegram", URL: "https://fake.test", Label: "link 1"},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "trims and persists",
			link:         models.ExtraLink{SectionID: extraSectionID(1), Type: "telegram", URL: "  https://fake.test  ", Label: "  link 1  "},
			resultWanted: &models.ExtraLink{SectionID: extraSectionID(1), Type: "telegram", URL: "https://fake.test", Label: "link 1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraLinkRepo{createErr: tt.createErr}
			s := NewExtraLinkService(repo)
			err := s.Create(context.Background(), tt.link)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				return
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createLink, *tt.resultWanted) {
				t.Fatalf("repo.Create: got %+v want %+v", repo.createLink, *tt.resultWanted)
			}
		})
	}
}

func TestExtraLinkService_Update(t *testing.T) {
	link := models.ExtraLink{Type: "telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil}

	tests := []struct {
		name         string
		idStr        string
		id           int
		link         models.ExtraLink
		updateErr    error
		err          error
		resultWanted *models.ExtraLink
	}{
		{
			name:  "reject non numerical id",
			idStr: "abc",
			link:  link,
			err:   errs.ErrExtraLinkInvalidID,
		},
		{
			name:  "reject id = 0",
			idStr: "0",
			link:  link,
			err:   errs.ErrExtraLinkInvalidID,
		},
		{
			name:      "repo update error",
			idStr:     "10",
			id:        10,
			link:      link,
			updateErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:         "accept a valid id and link",
			idStr:        "10",
			id:           10,
			link:         link,
			resultWanted: &link,
		},
		{
			name:         "accept a valid id with spaces",
			idStr:        "  10  ",
			id:           10,
			link:         link,
			resultWanted: &link,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraLinkRepo{updateErr: tt.updateErr}
			s := NewExtraLinkService(repo)
			err := s.Update(context.Background(), tt.link, tt.idStr)
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
			if !reflect.DeepEqual(repo.updateLink, *tt.resultWanted) {
				t.Fatalf("repo.Update: got %+v want %+v", repo.updateLink, *tt.resultWanted)
			}
		})
	}
}

func TestExtraLinkService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		idStr     string
		id        int
		err       error
		deleteErr error
	}{
		{name: "reject non numerical id", idStr: "abc", err: errs.ErrExtraLinkInvalidID},
		{name: "reject empty id with spaces", idStr: "  ", err: errs.ErrExtraLinkInvalidID},
		{name: "reject empty id without spaces", idStr: "", err: errs.ErrExtraLinkInvalidID},
		{name: "reject id = 0", idStr: "0", err: errs.ErrExtraLinkInvalidID},
		{name: "reject id < 0", idStr: "-10", err: errs.ErrExtraLinkInvalidID},
		{name: "repo delete error", idStr: "10", deleteErr: errs.ErrDatabaseDown, err: errs.ErrDatabaseDown},
		{name: "accept a valid id", idStr: "10", id: 10},
		{name: "accept a valid id with spaces", idStr: "  10  ", id: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExtraLinkRepo{deleteErr: tt.deleteErr}
			s := NewExtraLinkService(repo)
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
				t.Fatalf("Delete ExtraLinkService succeeded, want error %v", tt.err)
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
