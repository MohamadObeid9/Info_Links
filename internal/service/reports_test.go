package service

import (
	"context"
	"errors"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

// fakeReportRepo implements repository.ReportRepository for service tests.
// Add fields (e.g. listReturn, deleteErr) as you test List/Delete/Update.
type fakeReportRepo struct {
	createCalls int
	lastCreate  models.Report
	createErr   error
}

func (f *fakeReportRepo) Create(ctx context.Context, report models.Report) error {
	f.createCalls++
	f.lastCreate = report
	return f.createErr
}

func (f *fakeReportRepo) List(ctx context.Context, limit, offset int, q, status string) ([]models.Report, error) {
	return nil, nil
}

func (f *fakeReportRepo) Update(ctx context.Context, status string, id int) error {
	return nil
}

func (f *fakeReportRepo) Delete(ctx context.Context, id int) error {
	return nil
}

// TestReportService_Create shows the service pattern: inject fake repo,
// call the method, use errors.Is for sentinels and inspect fake state.
func TestReportService_Create(t *testing.T) {
	t.Run("rejects empty course or link without calling repo", func(t *testing.T) {
		repo := &fakeReportRepo{}
		svc := NewReportService(repo)

		err := svc.Create(context.Background(), models.Report{CourseName: "", LinkURL: "https://a"})
		if !errors.Is(err, errs.ErrCourseNameAndLinkUrlAreRequired) {
			t.Fatalf("err: got %v want %v", err, errs.ErrCourseNameAndLinkUrlAreRequired)
		}
		if repo.createCalls != 0 {
			t.Fatalf("repo.Create calls: got %d want 0", repo.createCalls)
		}
	})

	t.Run("trims fields and persists via repo", func(t *testing.T) {
		repo := &fakeReportRepo{}
		svc := NewReportService(repo)

		err := svc.Create(context.Background(), models.Report{
			CourseName:  "  My Course  ",
			LinkURL:     "  https://b  ",
			Description: "  note  ",
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if repo.createCalls != 1 {
			t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
		}
		if repo.lastCreate.CourseName != "My Course" || repo.lastCreate.LinkURL != "https://b" {
			t.Fatalf("unexpected trimmed values: %+v", repo.lastCreate)
		}
		if repo.lastCreate.Description != "note" {
			t.Fatalf("description: got %q want %q", repo.lastCreate.Description, "note")
		}
	})
}
