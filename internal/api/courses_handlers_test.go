package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeCourseService struct {
	createCalls  int
	createCourse models.Course
	createErr    error

	deleteCalls int
	deleteID    string
	deleteErr   error

	updateCalls  int
	updatePatch  models.CoursePatch
	updateID     string
	updateErr    error
}

func (f *fakeCourseService) Create(ctx context.Context, course models.Course) error {
	f.createCalls++
	f.createCourse = course
	return f.createErr
}

func (f *fakeCourseService) Delete(ctx context.Context, idStr string) error {
	f.deleteCalls++
	f.deleteID = idStr
	return f.deleteErr
}

func (f *fakeCourseService) Update(ctx context.Context, patch models.CoursePatch, idStr string) error {
	f.updateCalls++
	f.updatePatch = patch
	f.updateID = idStr
	return f.updateErr
}

func TestHandleAdminPostCourse(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		createErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		resultWanted *models.Course
	}{
		{
			name:         "201 when service accepts the course",
			body:         `{"name":"BDD","code":"nfa008","semester_id":3,"is_optional":false,"display_order":55}`,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
			resultWanted: &models.Course{Name: "BDD", Code: "nfa008", SemesterID: 3, IsOptional: false, DisplayOrder: 55},
		},
		{
			name:         "400 invalid JSON body",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 when service returns validation error",
			body:         `{"name":"","code":"nfa008","semester_id":3}`,
			createErr:    errs.ErrCourseCodeAndNameRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course code and course name are required",
			wantCalls:    1,
		},
		{
			name:         "400 when service returns invalid semester id",
			body:         `{"name":"BDD","code":"nfa008","semester_id":0}`,
			createErr:    errs.ErrCourseInvalidSemestreID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course invalid semestre id ",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			body:         `{"name":"BDD","code":"nfa008","semester_id":3}`,
			createErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCourseService{createErr: tt.createErr}
			h := testHandler(t, withCourse(fake))
			req := httptest.NewRequest(http.MethodPost, "/api/admin/courses", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPostCourse(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusCreated {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.createCalls != tt.wantCalls {
					t.Fatalf("service.Create calls: got %d want %d", fake.createCalls, tt.wantCalls)
				}
				return
			}

			if fake.createCalls != tt.wantCalls {
				t.Fatalf("service.Create calls: got %d want %d", fake.createCalls, tt.wantCalls)
			}
			if tt.resultWanted == nil {
				t.Fatal("success case must set resultWanted")
			}
			if !reflect.DeepEqual(fake.createCourse, *tt.resultWanted) {
				t.Fatalf("service.Create course: got %+v want %+v", fake.createCourse, *tt.resultWanted)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("status: got %q want %q", got["status"], "ok")
			}
		})
	}
}

func TestHandleAdminDeleteCourse(t *testing.T) {
	tests := []struct {
		name         string
		pathID       string
		deleteErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
	}{
		{
			name:         "400 invalid course id",
			pathID:       "abc",
			deleteErr:    errs.ErrCourseInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course invalid id",
			wantCalls:    1,
		},
		{
			name:         "404 course not found",
			pathID:       "10",
			deleteErr:    errs.ErrCourseNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Course not found",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			deleteErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept a valid id",
			pathID:       "10",
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
		},
		{
			name:         "accept a valid id with spaces",
			pathID:       "  10  ",
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "  10  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCourseService{deleteErr: tt.deleteErr}
			h := testHandler(t, withCourse(fake))
			req := httptest.NewRequest(http.MethodDelete, "/api/admin/courses/10", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.handleAdminDeleteCourse(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.deleteCalls != tt.wantCalls {
					t.Fatalf("service.Delete calls: got %d want %d", fake.deleteCalls, tt.wantCalls)
				}
				return
			}

			if fake.deleteCalls != tt.wantCalls {
				t.Fatalf("service.Delete calls: got %d want %d", fake.deleteCalls, tt.wantCalls)
			}
			if fake.deleteID != tt.wantID {
				t.Fatalf("service.Delete id: got %q want %q", fake.deleteID, tt.wantID)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("response status: got %q want ok", got["status"])
			}
		})
	}
}

func courseStrPtr(s string) *string { return &s }

func TestHandleAdminPatchCourse(t *testing.T) {
	tests := []struct {
		name         string
		pathID       string
		body         string
		updateErr    error
		statusWanted int
		errMsg       string
		wantCalls    int
		wantID       string
		patchWanted  *models.CoursePatch
	}{
		{
			name:         "400 invalid JSON body",
			pathID:       "10",
			body:         `{`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
			wantCalls:    0,
		},
		{
			name:         "400 invalid course id",
			pathID:       "abc",
			body:         `{"name":"Updated"}`,
			updateErr:    errs.ErrCourseInvalidID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course invalid id",
			wantCalls:    1,
		},
		{
			name:         "404 course not found",
			pathID:       "10",
			body:         `{"name":"Updated"}`,
			updateErr:    errs.ErrCourseNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "Course not found",
			wantCalls:    1,
		},
		{
			name:         "400 invalid semester id",
			pathID:       "10",
			body:         `{"semester_id":0}`,
			updateErr:    errs.ErrCourseInvalidSemestreID,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course invalid semestre id ",
			wantCalls:    1,
		},
		{
			name:         "400 empty patch",
			pathID:       "10",
			body:         `{}`,
			updateErr:    errs.ErrCoursePatchEmpty,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course invalid update parameters",
			wantCalls:    1,
		},
		{
			name:         "400 when name and code are required",
			pathID:       "10",
			body:         `{"name":"  "}`,
			updateErr:    errs.ErrCourseCodeAndNameRequired,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Course code and course name are required",
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			pathID:       "10",
			body:         `{"name":"Updated"}`,
			updateErr:    errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
		{
			name:         "accept valid patch",
			pathID:       "10",
			body:         `{"name":"Updated"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "10",
			patchWanted:  &models.CoursePatch{Name: courseStrPtr("Updated")},
		},
		{
			name:         "accept a valid id with spaces",
			pathID:       "  10  ",
			body:         `{"code":"nfa010"}`,
			statusWanted: http.StatusOK,
			wantCalls:    1,
			wantID:       "  10  ",
			patchWanted:  &models.CoursePatch{Code: courseStrPtr("nfa010")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCourseService{updateErr: tt.updateErr}
			h := testHandler(t, withCourse(fake))
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/courses/10", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.handleAdminPatchCourse(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status: got %d want %d body=%q", rr.Code, tt.statusWanted, rr.Body.String())
			}

			if tt.statusWanted != http.StatusOK {
				var got map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("json decode: %v", err)
				}
				if got["error"] != tt.errMsg {
					t.Fatalf("error: got %q want %q", got["error"], tt.errMsg)
				}
				if fake.updateCalls != tt.wantCalls {
					t.Fatalf("service.Update calls: got %d want %d", fake.updateCalls, tt.wantCalls)
				}
				return
			}

			if fake.updateCalls != tt.wantCalls {
				t.Fatalf("service.Update calls: got %d want %d", fake.updateCalls, tt.wantCalls)
			}
			if fake.updateID != tt.wantID {
				t.Fatalf("service.Update id: got %q want %q", fake.updateID, tt.wantID)
			}
			if tt.patchWanted == nil {
				t.Fatal("success case must set patchWanted")
			}
			if !reflect.DeepEqual(fake.updatePatch, *tt.patchWanted) {
				t.Fatalf("service.Update patch: got %+v want %+v", fake.updatePatch, *tt.patchWanted)
			}
			var got map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("json decode: %v", err)
			}
			if got["status"] != "ok" {
				t.Fatalf("response status: got %q want ok", got["status"])
			}
		})
	}
}
