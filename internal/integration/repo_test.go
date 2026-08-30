//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

func TestRepoUserGuestClaimAndCredentials(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	repo := newUserRepo(t, dbClient.DB)
	ctx := context.Background()

	guestID, err := repo.CreateGuest(ctx)
	if err != nil {
		t.Fatalf("CreateGuest: %v", err)
	}

	user, err := repo.ClaimGuest(ctx, guestID, models.User{
		FirstName: "ali",
		LastName:  "hassan",
		Number:    42,
	})
	if err != nil {
		t.Fatalf("ClaimGuest: %v", err)
	}
	if user.IsGuest {
		t.Fatal("claimed user should not be guest")
	}
	if user.ID != guestID {
		t.Fatalf("claim should keep id: got %d want %d", user.ID, guestID)
	}

	found, err := repo.GetByCredentials(ctx, models.User{
		FirstName: "ali",
		LastName:  "hassan",
		Number:    42,
	})
	if err != nil {
		t.Fatalf("GetByCredentials: %v", err)
	}
	if found.Handle != "ali_hassan_42" {
		t.Fatalf("handle: got %q want ali_hassan_42", found.Handle)
	}

	_, err = repo.ClaimGuest(ctx, guestID, models.User{
		FirstName: "other",
		LastName:  "person",
		Number:    1,
	})
	if err == nil {
		t.Fatal("expected error claiming already-registered row")
	}
	if err != errs.ErrUserGuestNotFound {
		t.Fatalf("re-claim error: got %v want ErrUserGuestNotFound", err)
	}
}

func TestRepoContentGetReturnsJSON(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	repo := newContentRepo(t, dbClient.DB)
	raw, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get content: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("content is not valid JSON: %v", err)
	}
	for _, key := range []string{"programs", "years", "courses", "links", "services"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("content JSON missing key %q", key)
		}
	}
}

func TestRepoServiceCreateAndList(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	repo := newServiceRepo(t, dbClient.DB)
	ctx := context.Background()

	expires := time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339)
	started := time.Now().Format(time.RFC3339)

	id, err := repo.Create(ctx, models.Service{
		Title:        "Tutoring",
		OwnerName:    "Ziad",
		Category:     "tutoring",
		Status:       "trial",
		Trial:        true,
		StartedAt:    started,
		ExpiresAt:    expires,
		DisplayOrder: 1,
		Links: []models.ServiceLink{
			{Label: "Telegram", URL: "https://t.me/example"},
		},
	})
	if err != nil {
		t.Fatalf("Create service: %v", err)
	}
	if id <= 0 {
		t.Fatalf("service id: got %d want > 0", id)
	}

	services, err := repo.List(ctx, 10, 0, "")
	if err != nil {
		t.Fatalf("List services: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("list len: got %d want 1", len(services))
	}
	if services[0].Title != "Tutoring" {
		t.Fatalf("title: got %q", services[0].Title)
	}
	if len(services[0].Links) != 1 || services[0].Links[0].URL != "https://t.me/example" {
		t.Fatalf("links: got %#v", services[0].Links)
	}
}
