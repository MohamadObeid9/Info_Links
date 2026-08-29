package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestServiceRepo(t *testing.T) (ServiceRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresServiceRepository(db), mock
}

func TestServiceRepository_Create(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	svc := models.Service{
		Title:       "Tutoring",
		OwnerName:   "Ali",
		Category:    "tutoring",
		Emoji:       "📚",
		Description: "Math help",
		LogoURL:     "https://example.com/logo.png",
		Phone:       "+96171123456",
		URL:         "https://t.me/ali",
		Links:       []models.ServiceLink{{Label: "Telegram", URL: "https://t.me/ali"}},
		Status:      "trial",
		Trial:       true,
		StartedAt:   "2026-08-27T00:00:00Z",
		ExpiresAt:   "2026-09-11T00:00:00Z",
	}

	mock.ExpectQuery(insertServiceQuery).
		WithArgs(svc.Title, svc.OwnerName, svc.Category, svc.Emoji, svc.Description, svc.LogoURL, svc.Phone, svc.URL, `[{"label":"Telegram","url":"https://t.me/ali"}]`, svc.Status, svc.Trial, svc.StartedAt, svc.ExpiresAt, svc.DisplayOrder).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := repo.Create(context.Background(), svc)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 1 {
		t.Fatalf("got id %d, want 1", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_List(t *testing.T) {
	columns := []string{"id", "title", "owner_name", "category", "emoji", "description", "logo_url", "phone", "url", "links", "status", "trial", "started_at", "expires_at", "display_order", "created_at", "updated_at", "clicks"}
	query := listServicesWithClicksQuery + ` ORDER BY s.display_order ASC, s.created_at DESC LIMIT $1 OFFSET $2`

	repo, mock := newTestServiceRepo(t)
	mock.ExpectQuery(query).
		WithArgs(25, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Tutoring", "Ali", "tutoring", "📚", "Math help", "", "+9617", "https://t.me/ali", "[]", "trial", true, "2026-08-27T00:00:00Z", "2026-09-11T00:00:00Z", 0, "2026-08-27T00:00:00Z", "2026-08-27T00:00:00Z", 5).
			AddRow(2, "Design", "Sam", "design", "🎨", "", "", "", "", "[]", "active", false, "2026-08-01T00:00:00Z", "2026-09-01T00:00:00Z", 1, "2026-08-01T00:00:00Z", "2026-08-01T00:00:00Z", 0))

	services, err := repo.List(context.Background(), 25, 0, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("got %d services, want 2", len(services))
	}
	if services[0].Clicks != 5 {
		t.Fatalf("got clicks %d, want 5", services[0].Clicks)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_List_WithSearch(t *testing.T) {
	columns := []string{"id", "title", "owner_name", "category", "emoji", "description", "logo_url", "phone", "url", "links", "status", "trial", "started_at", "expires_at", "display_order", "created_at", "updated_at", "clicks"}
	query := listServicesWithClicksQuery + listServicesSearchWhere + ` ORDER BY s.display_order ASC, s.created_at DESC LIMIT $3 OFFSET $4`

	repo, mock := newTestServiceRepo(t)
	mock.ExpectQuery(query).
		WithArgs("%ali%", "%ali%", 25, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Tutoring", "Ali", "tutoring", "📚", "Math", "", "", "", "[]", "trial", true, "2026-08-27T00:00:00Z", "2026-09-11T00:00:00Z", 0, "2026-08-27T00:00:00Z", "2026-08-27T00:00:00Z", 0))

	services, err := repo.List(context.Background(), 25, 0, "ali")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("got %d services, want 1", len(services))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_Get_NotFound(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectQuery(getServiceByIDQuery).WithArgs(99).WillReturnError(sql.ErrNoRows)

	_, err := repo.Get(context.Background(), 99)
	if !errors.Is(err, errs.ErrServiceNotFound) {
		t.Fatalf("got %v, want ErrServiceNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_Delete(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectExec(deleteServiceQuery).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_Renew(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectExec(renewServiceQuery).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Renew(context.Background(), 1); err != nil {
		t.Fatalf("Renew: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_FreezeExpired(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectExec(freezeExpiredServicesQuery).WillReturnResult(sqlmock.NewResult(0, 2))

	if err := repo.FreezeExpired(context.Background()); err != nil {
		t.Fatalf("FreezeExpired: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_SetStatus(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectExec(setServiceStatusQuery).WithArgs("frozen", 1).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.SetStatus(context.Background(), 1, "frozen"); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_InsertClick(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectExec(insertServiceClickQuery).
		WithArgs(1, 42, "home", "WhatsApp", "https://wa.me/9611234567", "phone").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.InsertClick(context.Background(), models.ServiceClick{
		ServiceID: 1, UserID: 42, PageContext: "home", LinkTarget: "WhatsApp",
		URL: "https://wa.me/9611234567", DeviceType: "phone",
	}); err != nil {
		t.Fatalf("InsertClick: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_Update(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	svc := models.Service{
		Title:       "Updated",
		OwnerName:   "Ali",
		Category:    "tutoring",
		Emoji:       "📚",
		Description: "Updated desc",
		LogoURL:     "https://example.com/a.png",
		Phone:       "71123456",
		URL:         "https://t.me/ali",
		Links:       []models.ServiceLink{{Label: "Tel", URL: "tel:123"}},
		Status:      "active",
		Trial:       false,
		StartedAt:   "2026-08-27T00:00:00Z",
		ExpiresAt:   "2026-09-27T00:00:00Z",
		DisplayOrder: 1,
	}
	mock.ExpectExec(updateServiceQuery).
		WithArgs(svc.Title, svc.OwnerName, svc.Category, svc.Emoji, svc.Description, svc.LogoURL, svc.Phone, svc.URL, `[{"label":"Tel","url":"tel:123"}]`, svc.Status, svc.Trial, svc.StartedAt, svc.ExpiresAt, svc.DisplayOrder, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), svc, 1); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestServiceRepository_GetClickCount(t *testing.T) {
	repo, mock := newTestServiceRepo(t)
	mock.ExpectQuery(countServiceClicksQuery).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	count, err := repo.GetClickCount(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetClickCount: %v", err)
	}
	if count != 7 {
		t.Fatalf("got %d, want 7", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNewService(t *testing.T) {
	svc := NewService("Tutoring", "Ali", "tutoring", "📚", "Math help", "", "71123456", "https://t.me/ali", []models.ServiceLink{{Label: "Tel", URL: "tel:123"}}, "trial", true, 15)
	if svc.Title != "Tutoring" || svc.Description != "Math help" || svc.Phone == "" {
		t.Fatalf("unexpected service: %+v", svc)
	}
	if svc.Status != "trial" || !svc.Trial {
		t.Fatalf("expected trial service")
	}
}
