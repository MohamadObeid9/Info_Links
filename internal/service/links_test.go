package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeLinkRepo struct {
	createCalls int
	createLink  models.Link
	createErr   error

	deleteCalls int
	deleteID    int
	deleteErr   error

	updateCalls int
	updateLink  models.Link
	updateID    int
	updateErr   error
}

func (f *fakeLinkRepo) Create(ctx context.Context, link models.Link) error {
	f.createCalls++
	f.createLink = link
	return f.createErr
}

func (f *fakeLinkRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeLinkRepo) Update(ctx context.Context, link models.Link, id int) error {
	f.updateCalls++
	f.updateLink = link
	f.updateID = id
	return f.updateErr
}

func TestLinkService_Create(t *testing.T) {
	tests := []struct {
		name         string
		link         models.Link
		createErr    error
		err          error
		resultWanted *models.Link
	}{
		{
			name: "label is required",
			link: models.Link{URL: "https://fake.test", Label: ""},
			err:  errs.ErrLinkURLAndLabelRequired,
		},
		{
			name: "url is required",
			link: models.Link{URL: "", Label: "link 1"},
			err:  errs.ErrLinkURLAndLabelRequired,
		},
		{
			name: "label shouldn't be empty",
			link: models.Link{URL: "https://fake.test", Label: "  "},
			err:  errs.ErrLinkURLAndLabelRequired,
		},
		{
			name: "url shouldn't be empty",
			link: models.Link{URL: "  ", Label: "link 1"},
			err:  errs.ErrLinkURLAndLabelRequired,
		},
		{
			name:      "repo create error",
			link:      models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", DisplayOrder: 5},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name: "trims and persists",
			link: models.Link{Type: "Telegram", URL: "  https://fake.test  ", Label: "  link 1  ", Note: "note", DisplayOrder: 5},
			resultWanted: &models.Link{
				Type:         "Telegram",
				URL:          "https://fake.test",
				Label:        "link 1",
				Note:         "note",
				DisplayOrder: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLinkRepo{createErr: tt.createErr}
			s := NewLinkService(repo)
			err := s.Create(context.Background(), tt.link)
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
				t.Fatalf("Create Link Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createLink, *tt.resultWanted) {
				t.Fatalf("repo.Create link: got %+v want %+v", repo.createLink, *tt.resultWanted)
			}
		})
	}
}

func TestLinkService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		idStr     string
		id        int
		err       error
		deleteErr error
	}{
		{
			name:  "reject non numerical id",
			idStr: "ABD",
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject empty id with spaces",
			idStr: "  ",
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject empty id without spaces",
			idStr: "",
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject id = 0",
			idStr: "0",
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject id < 0",
			idStr: "-10",
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:      "repo delete error",
			idStr:     "10",
			err:       errs.ErrDatabaseDown,
			deleteErr: errs.ErrDatabaseDown,
		},
		{
			name:  "accept a valid id",
			idStr: "10",
			id:    10,
		},
		{
			name:  "accept a valid id with spaces",
			idStr: "  10  ",
			id:    10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLinkRepo{deleteErr: tt.deleteErr}
			s := NewLinkService(repo)
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
				t.Fatalf("Delete Link Service succeeded, want error %v", tt.err)
			}
			if repo.deleteCalls != 1 {
				t.Fatalf("repo.Delete calls: got %d want 1", repo.deleteCalls)
			}
			if repo.deleteID != tt.id {
				t.Fatalf("repo.DeleteID: got %v, want %v", repo.deleteID, tt.id)
			}
		})
	}
}

func TestLinkService_Update(t *testing.T) {
	link := models.Link{Type: "Telegram", URL: "https://fake.test", Label: "link 1", Note: "", ContentType: nil}

	tests := []struct {
		name         string
		idStr        string
		id           int
		link         models.Link
		updateErr    error
		err          error
		resultWanted *models.Link
	}{
		{
			name:  "reject non numerical id",
			idStr: "abc",
			link:  link,
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject empty id with spaces",
			idStr: "  ",
			link:  link,
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject empty id without spaces",
			idStr: "",
			link:  link,
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject id = 0",
			idStr: "0",
			link:  link,
			err:   errs.ErrLinkInvalidID,
		},
		{
			name:  "reject id < 0",
			idStr: "-10",
			link:  link,
			err:   errs.ErrLinkInvalidID,
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
			repo := &fakeLinkRepo{updateErr: tt.updateErr}
			s := NewLinkService(repo)
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
			if tt.err != nil {
				t.Fatalf("Update Link Service succeeded, want error %v", tt.err)
			}
			if repo.updateCalls != 1 {
				t.Fatalf("repo.Update calls: got %d want 1", repo.updateCalls)
			}
			if repo.updateID != tt.id {
				t.Fatalf("repo.Update id: got %d want %d", repo.updateID, tt.id)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.updateLink, *tt.resultWanted) {
				t.Fatalf("repo.Update link: got %+v want %+v", repo.updateLink, *tt.resultWanted)
			}
		})
	}
}
