package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeCourseRepo struct {
	createCalls  int
	createCourse models.Course
	createErr    error

	deleteCalls       int
	deleteID          int
	deletePlacementID int
	deleteErr         error

	getByIDCalls  int
	getByID       int
	getByIDErr    error
	getByIDResult models.Course

	updateCalls  int
	updateCourse models.Course
	updateID     int
	updateErr    error
}

func (f *fakeCourseRepo) Create(ctx context.Context, course models.Course) error {
	f.createCalls++
	f.createCourse = course
	return f.createErr
}

func (f *fakeCourseRepo) Delete(ctx context.Context, id int) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeCourseRepo) DeletePlacement(ctx context.Context, courseID, placementID int) error {
	f.deleteCalls++
	f.deleteID = courseID
	f.deletePlacementID = placementID
	return f.deleteErr
}

func (f *fakeCourseRepo) GetByID(ctx context.Context, id int) (models.Course, error) {
	f.getByIDCalls++
	f.getByID = id
	if f.getByIDErr != nil {
		return models.Course{}, f.getByIDErr
	}
	return f.getByIDResult, nil
}

func (f *fakeCourseRepo) Update(ctx context.Context, course models.Course, id int) error {
	f.updateCalls++
	f.updateCourse = course
	f.updateID = id
	return f.updateErr
}

func TestCourseService_Create(t *testing.T) {
	tests := []struct {
		name         string
		course       models.Course
		createErr    error
		err          error
		resultWanted *models.Course
	}{
		{
			name:   "name is required",
			course: models.Course{Name: "", Code: "nfa008", SemesterID: 3},
			err:    errs.ErrCourseCodeAndNameRequired,
		},
		{
			name:   "code is required",
			course: models.Course{Name: "BDD", Code: "", SemesterID: 3},
			err:    errs.ErrCourseCodeAndNameRequired,
		},
		{
			name:   "name shouldn't be empty",
			course: models.Course{Name: "  ", Code: "nfa008", SemesterID: 3},
			err:    errs.ErrCourseCodeAndNameRequired,
		},
		{
			name:   "code shouldn't be empty",
			course: models.Course{Name: "BDD", Code: "  ", SemesterID: 3},
			err:    errs.ErrCourseCodeAndNameRequired,
		},
		{
			name:   "reject semester_id = 0",
			course: models.Course{Name: "BDD", Code: "nfa008", SemesterID: 0},
			err:    errs.ErrCourseInvalidSemestreID,
		},
		{
			name:   "reject semester_id < 0",
			course: models.Course{Name: "BDD", Code: "nfa008", SemesterID: -1},
			err:    errs.ErrCourseInvalidSemestreID,
		},
		{
			name:      "repo create error",
			course:    models.Course{Name: "BDD", Code: "nfa008", SemesterID: 3, IsOptional: false, DisplayOrder: 55},
			createErr: errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
		{
			name:   "trims and persists",
			course: models.Course{Name: "  BDD  ", Code: "  nfa008  ", SemesterID: 3, IsOptional: true, DisplayOrder: 55},
			resultWanted: &models.Course{
				Name:         "BDD",
				Code:         "nfa008",
				SemesterID:   3,
				IsOptional:   true,
				DisplayOrder: 55,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCourseRepo{createErr: tt.createErr}
			s := NewCourseService(repo)
			err := s.Create(context.Background(), tt.course)
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
				t.Fatalf("Create Course Service succeeded, want error: %v", tt.err)
			}
			if repo.createCalls != 1 {
				t.Fatalf("repo.Create calls: got %d want 1", repo.createCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.createCourse, *tt.resultWanted) {
				t.Fatalf("repo.Create course: got %+v want %+v", repo.createCourse, *tt.resultWanted)
			}
		})
	}
}

func TestCourseService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		idStr         string
		placementStr  string
		id            int
		wantPlacement int
		err           error
		deleteErr     error
	}{
		{
			name:  "reject non numerical id",
			idStr: "ABD",
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject empty id with spaces",
			idStr: "  ",
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject empty id without spaces",
			idStr: "",
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject id = 0",
			idStr: "0",
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject id < 0",
			idStr: "-10",
			err:   errs.ErrCourseInvalidID,
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
		{
			name:          "deletes a placement when placement_id is set",
			idStr:         "10",
			placementStr:  "7",
			id:            10,
			wantPlacement: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCourseRepo{deleteErr: tt.deleteErr}
			s := NewCourseService(repo)
			err := s.Delete(context.Background(), tt.idStr, tt.placementStr)
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
				t.Fatalf("Delete Course Service succeeded, want error %v", tt.err)
			}
			if repo.deleteCalls != 1 {
				t.Fatalf("repo.Delete calls: got %d want 1", repo.deleteCalls)
			}
			if repo.deleteID != tt.id {
				t.Fatalf("repo.DeleteID: got %v, want %v", repo.deleteID, tt.id)
			}
			if repo.deletePlacementID != tt.wantPlacement {
				t.Fatalf("repo.DeletePlacementID: got %v, want %v", repo.deletePlacementID, tt.wantPlacement)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

func TestCourseService_Update(t *testing.T) {
	existing := models.Course{
		ID:           10,
		SemesterID:   3,
		Name:         "Reseaux",
		Code:         "NFA009",
		IsOptional:   false,
		DisplayOrder: 55,
	}

	tests := []struct {
		name         string
		idStr        string
		id           int
		patch        models.CoursePatch
		existing     models.Course
		getByIDErr   error
		updateErr    error
		err          error
		resultWanted *models.Course
		wantGetByID  int
		wantUpdate   int
	}{
		{
			name:  "reject non numerical id",
			idStr: "abc",
			patch: models.CoursePatch{Name: strPtr("Updated")},
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject empty id with spaces",
			idStr: "  ",
			patch: models.CoursePatch{Name: strPtr("Updated")},
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject empty id without spaces",
			idStr: "",
			patch: models.CoursePatch{Name: strPtr("Updated")},
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject id = 0",
			idStr: "0",
			patch: models.CoursePatch{Name: strPtr("Updated")},
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:  "reject id < 0",
			idStr: "-10",
			patch: models.CoursePatch{Name: strPtr("Updated")},
			err:   errs.ErrCourseInvalidID,
		},
		{
			name:        "repo get by id error",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{Name: strPtr("Updated")},
			existing:    existing,
			getByIDErr:  errs.ErrCourseNotFound,
			err:         errs.ErrCourseNotFound,
			wantGetByID: 1,
		},
		{
			name:        "reject empty patch",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{},
			existing:    existing,
			err:         errs.ErrCoursePatchEmpty,
			wantGetByID: 1,
		},
		{
			name:        "reject semester_id = 0",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{SemesterID: intPtr(0)},
			existing:    existing,
			err:         errs.ErrCourseInvalidSemestreID,
			wantGetByID: 1,
		},
		{
			name:        "reject semester_id < 0",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{SemesterID: intPtr(-1)},
			existing:    existing,
			err:         errs.ErrCourseInvalidSemestreID,
			wantGetByID: 1,
		},
		{
			name:        "reject semester_id without placement_id",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{SemesterID: intPtr(5)},
			existing:    existing,
			err:         errs.ErrCourseInvalidPlacementID,
			wantGetByID: 1,
		},
		{
			name:        "reject empty name with spaces",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{Name: strPtr("  ")},
			existing:    existing,
			err:         errs.ErrCourseCodeAndNameRequired,
			wantGetByID: 1,
		},
		{
			name:        "reject empty code with spaces",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{Code: strPtr("  ")},
			existing:    existing,
			err:         errs.ErrCourseCodeAndNameRequired,
			wantGetByID: 1,
		},
		{
			name:        "repo update error",
			idStr:       "10",
			id:          10,
			patch:       models.CoursePatch{Name: strPtr("Updated")},
			existing:    existing,
			updateErr:   errs.ErrDatabaseDown,
			err:         errs.ErrDatabaseDown,
			wantGetByID: 1,
			wantUpdate:  1,
		},
		{
			name:     "accept patch name",
			idStr:    "10",
			id:       10,
			patch:    models.CoursePatch{Name: strPtr("  Reseaux 2  ")},
			existing: existing,
			resultWanted: &models.Course{
				ID:           10,
				SemesterID:   3,
				Name:         "Reseaux 2",
				Code:         "NFA009",
				IsOptional:   false,
				DisplayOrder: 55,
			},
			wantGetByID: 1,
			wantUpdate:  1,
		},
		{
			name:     "accept patch code",
			idStr:    "10",
			id:       10,
			patch:    models.CoursePatch{Code: strPtr("  nfa010  ")},
			existing: existing,
			resultWanted: &models.Course{
				ID:           10,
				SemesterID:   3,
				Name:         "Reseaux",
				Code:         "nfa010",
				IsOptional:   false,
				DisplayOrder: 55,
			},
			wantGetByID: 1,
			wantUpdate:  1,
		},
		{
			name:     "accept patch semester_id",
			idStr:    "10",
			id:       10,
			patch:    models.CoursePatch{SemesterID: intPtr(5), PlacementID: intPtr(9)},
			existing: existing,
			resultWanted: &models.Course{
				ID:           10,
				SemesterID:   5,
				PlacementID:  9,
				Name:         "Reseaux",
				Code:         "NFA009",
				IsOptional:   false,
				DisplayOrder: 55,
			},
			wantGetByID: 1,
			wantUpdate:  1,
		},
		{
			name:     "accept patch is_optional",
			idStr:    "10",
			id:       10,
			patch:    models.CoursePatch{IsOptional: boolPtr(true)},
			existing: existing,
			resultWanted: &models.Course{
				ID:           10,
				SemesterID:   3,
				Name:         "Reseaux",
				Code:         "NFA009",
				IsOptional:   true,
				DisplayOrder: 55,
			},
			wantGetByID: 1,
			wantUpdate:  1,
		},
		{
			name:     "accept a valid id with spaces",
			idStr:    "  10  ",
			id:       10,
			patch:    models.CoursePatch{Name: strPtr("Updated")},
			existing: existing,
			resultWanted: &models.Course{
				ID:           10,
				SemesterID:   3,
				Name:         "Updated",
				Code:         "NFA009",
				IsOptional:   false,
				DisplayOrder: 55,
			},
			wantGetByID: 1,
			wantUpdate:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCourseRepo{
				getByIDResult: tt.existing,
				getByIDErr:    tt.getByIDErr,
				updateErr:     tt.updateErr,
			}
			s := NewCourseService(repo)
			err := s.Update(context.Background(), tt.patch, tt.idStr)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("got %v, want %v", err, tt.err)
				}
				if repo.getByIDCalls != tt.wantGetByID {
					t.Fatalf("repo.GetByID calls: got %d want %d", repo.getByIDCalls, tt.wantGetByID)
				}
				if repo.updateCalls != tt.wantUpdate {
					t.Fatalf("repo.Update calls: got %d want %d", repo.updateCalls, tt.wantUpdate)
				}
				if tt.wantGetByID > 0 && repo.getByID != tt.id {
					t.Fatalf("repo.GetByID id: got %d want %d", repo.getByID, tt.id)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("Update Course Service succeeded, want error %v", tt.err)
			}
			if repo.getByIDCalls != tt.wantGetByID {
				t.Fatalf("repo.GetByID calls: got %d want %d", repo.getByIDCalls, tt.wantGetByID)
			}
			if repo.updateCalls != tt.wantUpdate {
				t.Fatalf("repo.Update calls: got %d want %d", repo.updateCalls, tt.wantUpdate)
			}
			if repo.getByID != tt.id {
				t.Fatalf("repo.GetByID id: got %d want %d", repo.getByID, tt.id)
			}
			if repo.updateID != tt.id {
				t.Fatalf("repo.Update id: got %d want %d", repo.updateID, tt.id)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(repo.updateCourse, *tt.resultWanted) {
				t.Fatalf("repo.Update course: got %+v want %+v", repo.updateCourse, *tt.resultWanted)
			}
		})
	}
}
