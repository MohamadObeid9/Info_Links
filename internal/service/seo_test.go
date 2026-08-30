package service

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/repository"
)

type fakeSEORepo struct {
	getCourseCalls int
	getCourseCode  string
	getCourseData  *repository.CoursePageData
	getCourseErr   error

	listCodesCalls int
	listCodes      []string
	listCodesErr   error

	listProgramsCalls int
	listPrograms      []repository.ProgramSitemapEntry
	listProgramsErr   error

	listIndexCalls int
	listIndex      []repository.CourseIndexEntry
	listIndexErr   error

	getProgramCalls int
	getProgramSlug  string
	getProgramData  *repository.ProgramPageData
	getProgramErr   error
}

func (f *fakeSEORepo) GetCoursePageByCode(ctx context.Context, code string) (*repository.CoursePageData, error) {
	f.getCourseCalls++
	f.getCourseCode = code
	if f.getCourseErr != nil {
		return nil, f.getCourseErr
	}
	return f.getCourseData, nil
}

func (f *fakeSEORepo) ListCourseCodesForSitemap(ctx context.Context) ([]string, error) {
	f.listCodesCalls++
	if f.listCodesErr != nil {
		return nil, f.listCodesErr
	}
	return f.listCodes, nil
}

func (f *fakeSEORepo) ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]repository.ProgramSitemapEntry, error) {
	f.listProgramsCalls++
	if f.listProgramsErr != nil {
		return nil, f.listProgramsErr
	}
	return f.listPrograms, nil
}

func (f *fakeSEORepo) ListCoursesIndex(ctx context.Context) ([]repository.CourseIndexEntry, error) {
	f.listIndexCalls++
	if f.listIndexErr != nil {
		return nil, f.listIndexErr
	}
	return f.listIndex, nil
}

func (f *fakeSEORepo) GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*repository.ProgramPageData, error) {
	f.getProgramCalls++
	f.getProgramSlug = slug
	if f.getProgramErr != nil {
		return nil, f.getProgramErr
	}
	return f.getProgramData, nil
}

func TestSEOService_GetCoursePageByCode(t *testing.T) {
	sample := &repository.CoursePageData{Code: "nfa008", Name: "BDD"}

	tests := []struct {
		name      string
		code      string
		repoData  *repository.CoursePageData
		repoErr   error
		want      *repository.CoursePageData
		wantErr   error
		wantCalls int
	}{
		{
			name:      "returns repo result",
			code:      "nfa008",
			repoData:  sample,
			want:      sample,
			wantCalls: 1,
		},
		{
			name:      "wraps repo error",
			code:      "nfa008",
			repoErr:   errs.ErrCourseNotFound,
			wantErr:   errs.ErrCourseNotFound,
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeSEORepo{getCourseData: tt.repoData, getCourseErr: tt.repoErr}
			svc := NewSEOService(repo)

			got, err := svc.GetCoursePageByCode(context.Background(), tt.code)
			if repo.getCourseCalls != tt.wantCalls {
				t.Fatalf("repo calls: got %d want %d", repo.getCourseCalls, tt.wantCalls)
			}
			if repo.getCourseCode != tt.code {
				t.Fatalf("repo code: got %q want %q", repo.getCourseCode, tt.code)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEOService_ListCourseCodesForSitemap(t *testing.T) {
	repo := &fakeSEORepo{listCodes: []string{"nfa008", "nfa010"}}
	svc := NewSEOService(repo)

	got, err := svc.ListCourseCodesForSitemap(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listCodesCalls != 1 {
		t.Fatalf("repo calls: got %d want 1", repo.listCodesCalls)
	}
	if !reflect.DeepEqual(got, repo.listCodes) {
		t.Fatalf("got %+v, want %+v", got, repo.listCodes)
	}
}

func TestSEOService_ListProgramsForSitemap(t *testing.T) {
	programs := []repository.ProgramSitemapEntry{{ID: 1, Name: "Génie Info", Slug: "genie-info"}}
	repo := &fakeSEORepo{listPrograms: programs}
	svc := NewSEOService(repo)

	got, err := svc.ListProgramsForSitemap(context.Background(), func(string) string { return "genie-info" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listProgramsCalls != 1 {
		t.Fatalf("repo calls: got %d want 1", repo.listProgramsCalls)
	}
	if !reflect.DeepEqual(got, programs) {
		t.Fatalf("got %+v, want %+v", got, programs)
	}
}

func TestSEOService_ListCoursesIndex(t *testing.T) {
	entries := []repository.CourseIndexEntry{{Code: "nfa008", Name: "BDD", ProgramName: "Génie Info"}}
	repo := &fakeSEORepo{listIndex: entries}
	svc := NewSEOService(repo)

	got, err := svc.ListCoursesIndex(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listIndexCalls != 1 {
		t.Fatalf("repo calls: got %d want 1", repo.listIndexCalls)
	}
	if !reflect.DeepEqual(got, entries) {
		t.Fatalf("got %+v, want %+v", got, entries)
	}
}

func TestSEOService_GetProgramBySlug(t *testing.T) {
	sample := &repository.ProgramPageData{ID: 1, Name: "Génie Info", Slug: "genie-info"}

	tests := []struct {
		name     string
		slug     string
		repoData *repository.ProgramPageData
		repoErr  error
		want     *repository.ProgramPageData
		wantErr  error
	}{
		{
			name:     "returns repo result",
			slug:     "genie-info",
			repoData: sample,
			want:     sample,
		},
		{
			name:    "wraps repo error",
			slug:    "missing",
			repoErr: sql.ErrNoRows,
			wantErr: sql.ErrNoRows,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeSEORepo{getProgramData: tt.repoData, getProgramErr: tt.repoErr}
			svc := NewSEOService(repo)

			got, err := svc.GetProgramBySlug(context.Background(), tt.slug, func(string) string { return tt.slug })
			if repo.getProgramSlug != tt.slug {
				t.Fatalf("repo slug: got %q want %q", repo.getProgramSlug, tt.slug)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSEOService_wrapsDatabaseErrors(t *testing.T) {
	repo := &fakeSEORepo{listCodesErr: errs.ErrDatabaseDown}
	svc := NewSEOService(repo)

	_, err := svc.ListCourseCodesForSitemap(context.Background())
	if !errors.Is(err, errs.ErrDatabaseDown) {
		t.Fatalf("got %v, want %v", err, errs.ErrDatabaseDown)
	}
}
