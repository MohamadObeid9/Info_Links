package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type fakeUserRepo struct {
	getCalls  int
	getResult int
	getErr    error

	claimCalls   int
	claimGuestID int
	claimUser    models.User
	claimResult  models.User
	claimErr     error

	createCalls  int
	createUser   models.User
	createResult models.User
	createErr    error

	credentialsCalls  int
	credentialsUser   models.User
	credentialsResult models.User
	credentialsErr    error

	adoptCalls   int
	adoptGuestID int
	adoptUserID  int
	adoptErr     error

	staleCalls  int
	staleCutoff time.Time
	staleResult int64
	staleErr    error

	byIDCalls  int
	byIDResult models.User
	byIDErr    error

	favoriteCalls    int
	favoriteUserID   int
	favoriteCourseID int
	favoriteErr      error

	studentsCalls  int
	studentsResult []models.UserListItem
	studentsErr    error

	activityCalls  int
	activityResult []models.UserActivityEvent
	activityErr    error

	lastDeviceCalls  int
	lastDeviceResult string
	lastDeviceErr    error
}

func (f *fakeUserRepo) CreateGuest(ctx context.Context) (int, error) {
	f.getCalls++
	if f.getErr != nil {
		return 0, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeUserRepo) ClaimGuest(ctx context.Context, guestID int, u models.User) (models.User, error) {
	f.claimCalls++
	f.claimGuestID = guestID
	f.claimUser = u
	if f.claimErr != nil {
		return models.User{}, f.claimErr
	}
	return f.claimResult, nil
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, u models.User) (models.User, error) {
	f.createCalls++
	f.createUser = u
	if f.createErr != nil {
		return models.User{}, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id int) (models.User, error) {
	f.byIDCalls++
	if f.byIDErr != nil {
		return models.User{}, f.byIDErr
	}
	return f.byIDResult, nil
}

func (f *fakeUserRepo) GetByCredentials(ctx context.Context, u models.User) (models.User, error) {
	f.credentialsCalls++
	f.credentialsUser = u
	if f.credentialsErr != nil {
		return models.User{}, f.credentialsErr
	}
	return f.credentialsResult, nil
}

func (f *fakeUserRepo) AdoptGuest(ctx context.Context, guestID int, userID int) error {
	f.adoptCalls++
	f.adoptGuestID = guestID
	f.adoptUserID = userID
	return f.adoptErr
}

func (f *fakeUserRepo) DeleteStaleGuests(ctx context.Context, olderThan time.Time) (int64, error) {
	f.staleCalls++
	f.staleCutoff = olderThan
	if f.staleErr != nil {
		return 0, f.staleErr
	}
	return f.staleResult, nil
}

func (f *fakeUserRepo) AddFavorite(ctx context.Context, userID int, courseID int) error {
	f.favoriteCalls++
	f.favoriteUserID = userID
	f.favoriteCourseID = courseID
	return f.favoriteErr
}

func (f *fakeUserRepo) RemoveFavorite(ctx context.Context, userID int, courseID int) error {
	f.favoriteCalls++
	f.favoriteUserID = userID
	f.favoriteCourseID = courseID
	return f.favoriteErr
}

func (f *fakeUserRepo) ListStudents(ctx context.Context, limit int, offset int, q string) ([]models.UserListItem, error) {
	f.studentsCalls++
	if f.studentsErr != nil {
		return nil, f.studentsErr
	}
	return f.studentsResult, nil
}

func (f *fakeUserRepo) ListActivity(ctx context.Context, userID int, limit int, offset int) ([]models.UserActivityEvent, error) {
	f.activityCalls++
	if f.activityErr != nil {
		return nil, f.activityErr
	}
	return f.activityResult, nil
}

func (f *fakeUserRepo) GetLastDeviceType(ctx context.Context, userID int) (string, error) {
	f.lastDeviceCalls++
	if f.lastDeviceErr != nil {
		return "", f.lastDeviceErr
	}
	return f.lastDeviceResult, nil
}

func TestUserService_CreateGuest(t *testing.T) {

	tests := []struct {
		name       string
		repoResult int
		repoErr    error
		wantResult int
		wantErr    error
		wantCalls  int
	}{
		{
			name:       "returns repo result",
			repoResult: 9,
			wantResult: 9,
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
			repo := &fakeUserRepo{getResult: tt.repoResult, getErr: tt.repoErr}
			svc := NewUserService(repo)

			got, err := svc.CreateGuest(context.Background())
			if repo.getCalls != tt.wantCalls {
				t.Fatalf("repo get calls = %d, want %d", repo.getCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Fatalf("got %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestUserService_RegisterUser(t *testing.T) {
	student := models.User{ID: 19, FirstName: "mohamad", LastName: "hassan", Number: 55}

	tests := []struct {
		name           string
		guestID        int
		input          models.User
		claimResult    models.User
		claimErr       error
		createResult   models.User
		createErr      error
		wantClaimCalls int
		wantCreateCall int
		wantRepoUser   models.User
		want           models.User
		wantErr        error
	}{
		{
			name:           "claims the guest and keeps its id",
			guestID:        19,
			input:          models.User{FirstName: "Mohamad", LastName: "Hassan", Number: 55},
			claimResult:    student,
			wantClaimCalls: 1,
			wantRepoUser:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:           student,
		},
		{
			name:           "creates a student when no guest token is sent",
			input:          models.User{FirstName: "  Mohamad ", LastName: " HASSAN", Number: 55},
			createResult:   student,
			wantCreateCall: 1,
			wantRepoUser:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:           student,
		},
		{
			name:           "falls through to a fresh student when the guest id is stale",
			guestID:        404,
			input:          models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			claimErr:       errs.ErrUserGuestNotFound,
			createResult:   student,
			wantClaimCalls: 1,
			wantCreateCall: 1,
			wantRepoUser:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:           student,
		},
		{
			name:           "propagates a taken name from the claim",
			guestID:        19,
			input:          models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			claimErr:       errs.ErrUsernameTaken,
			wantClaimCalls: 1,
			wantErr:        errs.ErrUsernameTaken,
		},
		{
			name:           "propagates a taken name from the create",
			input:          models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			createErr:      errs.ErrUsernameTaken,
			wantCreateCall: 1,
			wantErr:        errs.ErrUsernameTaken,
		},
		{
			name:    "rejects a blank name",
			input:   models.User{FirstName: "   ", LastName: "hassan", Number: 55},
			wantErr: errs.ErrUserNameRequired,
		},
		{
			name:    "rejects a blank last name",
			input:   models.User{FirstName: "mohamad", LastName: "", Number: 55},
			wantErr: errs.ErrUserNameRequired,
		},
		{
			name:    "rejects a missing number",
			input:   models.User{FirstName: "mohamad", LastName: "hassan"},
			wantErr: errs.ErrUserNumberRange,
		},
		{
			name:    "rejects a number above 100",
			input:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 101},
			wantErr: errs.ErrUserNumberRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{
				claimResult:  tt.claimResult,
				claimErr:     tt.claimErr,
				createResult: tt.createResult,
				createErr:    tt.createErr,
			}
			svc := NewUserService(repo)

			got, err := svc.RegisterUser(context.Background(), tt.guestID, tt.input)

			if repo.claimCalls != tt.wantClaimCalls {
				t.Fatalf("claim calls = %d, want %d", repo.claimCalls, tt.wantClaimCalls)
			}
			if repo.createCalls != tt.wantCreateCall {
				t.Fatalf("create calls = %d, want %d", repo.createCalls, tt.wantCreateCall)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RegisterUser: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if tt.wantClaimCalls > 0 {
				if repo.claimGuestID != tt.guestID {
					t.Fatalf("claim guest id = %d, want %d", repo.claimGuestID, tt.guestID)
				}
				if !reflect.DeepEqual(repo.claimUser, tt.wantRepoUser) {
					t.Fatalf("claim user = %+v, want %+v", repo.claimUser, tt.wantRepoUser)
				}
			}
			if tt.wantCreateCall > 0 && !reflect.DeepEqual(repo.createUser, tt.wantRepoUser) {
				t.Fatalf("create user = %+v, want %+v", repo.createUser, tt.wantRepoUser)
			}
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	student := models.User{ID: 19, FirstName: "mohamad", LastName: "hassan", Number: 55}

	tests := []struct {
		name           string
		guestID        int
		input          models.User
		repoResult     models.User
		repoErr        error
		adoptErr       error
		wantCalls      int
		wantAdoptCalls int
		wantRepoUser   models.User
		want           models.User
		wantErr        error
	}{
		{
			name:         "normalizes the credentials before the lookup",
			input:        models.User{FirstName: " Mohamad", LastName: "Hassan ", Number: 55},
			repoResult:   student,
			wantCalls:    1,
			wantRepoUser: models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:         student,
		},
		{
			name:           "moves the guest visit onto the signed-in student",
			guestID:        42,
			input:          models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			repoResult:     student,
			wantCalls:      1,
			wantAdoptCalls: 1,
			wantRepoUser:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:           student,
		},
		{
			name:           "login still succeeds when the guest row is already gone",
			guestID:        42,
			input:          models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			repoResult:     student,
			adoptErr:       errs.ErrUserGuestNotFound,
			wantCalls:      1,
			wantAdoptCalls: 1,
			wantRepoUser:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:           student,
		},
		{
			name:         "does not adopt when the lookup fails",
			guestID:      42,
			input:        models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			repoErr:      errs.ErrUserNotFound,
			wantCalls:    1,
			wantRepoUser: models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			wantErr:      errs.ErrUserNotFound,
		},
		{
			name:         "skips adopt when the guest id is already the student",
			guestID:      19,
			input:        models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			repoResult:   student,
			wantCalls:    1,
			wantRepoUser: models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			want:         student,
		},
		{
			name:      "propagates an unknown student",
			input:     models.User{FirstName: "mohamad", LastName: "hassan", Number: 55},
			repoErr:   errs.ErrUserNotFound,
			wantCalls: 1,
			wantErr:   errs.ErrUserNotFound,
		},
		{
			name:    "rejects a blank name",
			input:   models.User{Number: 55},
			wantErr: errs.ErrUserNameRequired,
		},
		{
			name:    "rejects a number out of range",
			input:   models.User{FirstName: "mohamad", LastName: "hassan", Number: 0},
			wantErr: errs.ErrUserNumberRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{credentialsResult: tt.repoResult, credentialsErr: tt.repoErr, adoptErr: tt.adoptErr}
			svc := NewUserService(repo)

			got, err := svc.LoginUser(context.Background(), tt.guestID, tt.input)

			if repo.credentialsCalls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.credentialsCalls, tt.wantCalls)
			}
			if repo.adoptCalls != tt.wantAdoptCalls {
				t.Fatalf("adopt calls = %d, want %d", repo.adoptCalls, tt.wantAdoptCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoginUser: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if !reflect.DeepEqual(repo.credentialsUser, tt.wantRepoUser) {
				t.Fatalf("repo user = %+v, want %+v", repo.credentialsUser, tt.wantRepoUser)
			}
			if tt.wantAdoptCalls > 0 {
				if repo.adoptGuestID != tt.guestID || repo.adoptUserID != student.ID {
					t.Fatalf("adopt(%d, %d), want (%d, %d)", repo.adoptGuestID, repo.adoptUserID, tt.guestID, student.ID)
				}
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	student := models.User{ID: 19, FirstName: "mohamad", LastName: "hassan", Number: 55}

	tests := []struct {
		name       string
		userID     int
		repoResult models.User
		repoErr    error
		wantCalls  int
		want       models.User
		wantErr    error
	}{
		{
			name:       "returns the student",
			userID:     19,
			repoResult: student,
			wantCalls:  1,
			want:       student,
		},
		{
			name:      "propagates a missing student",
			userID:    19,
			repoErr:   errs.ErrUserNotFound,
			wantCalls: 1,
			wantErr:   errs.ErrUserNotFound,
		},
		{
			name:    "rejects a non positive id",
			userID:  0,
			wantErr: errs.ErrUserInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{byIDResult: tt.repoResult, byIDErr: tt.repoErr}
			svc := NewUserService(repo)

			got, err := svc.GetUser(context.Background(), tt.userID)

			if repo.byIDCalls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.byIDCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUser: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserService_Favorites(t *testing.T) {
	tests := []struct {
		name         string
		remove       bool
		courseIDStr  string
		repoErr      error
		wantCalls    int
		wantCourseID int
		wantErr      error
	}{
		{
			name:         "adds a favorite",
			courseIDStr:  " 3 ",
			wantCalls:    1,
			wantCourseID: 3,
		},
		{
			name:         "removes a favorite",
			remove:       true,
			courseIDStr:  "3",
			wantCalls:    1,
			wantCourseID: 3,
		},
		{
			name:        "rejects a non numeric course id",
			courseIDStr: "abc",
			wantErr:     errs.ErrCourseInvalidID,
		},
		{
			name:        "rejects a non positive course id",
			courseIDStr: "0",
			wantErr:     errs.ErrCourseInvalidID,
		},
		{
			name:        "propagates an unknown course",
			courseIDStr: "3",
			repoErr:     errs.ErrCourseNotFound,
			wantCalls:   1,
			wantErr:     errs.ErrCourseNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{favoriteErr: tt.repoErr}
			svc := NewUserService(repo)

			var err error
			if tt.remove {
				err = svc.RemoveFavorite(context.Background(), 19, tt.courseIDStr)
			} else {
				err = svc.AddFavorite(context.Background(), 19, tt.courseIDStr)
			}

			if repo.favoriteCalls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.favoriteCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("favorite: %v", err)
			}
			if repo.favoriteUserID != 19 || repo.favoriteCourseID != tt.wantCourseID {
				t.Fatalf("repo args = (%d, %d), want (19, %d)", repo.favoriteUserID, repo.favoriteCourseID, tt.wantCourseID)
			}
		})
	}
}

func TestUserService_ListStudents(t *testing.T) {
	students := []models.UserListItem{{ID: 1, Handle: "mohamad_hassan_55"}}

	tests := []struct {
		name       string
		limit      int
		offset     int
		repoResult []models.UserListItem
		repoErr    error
		wantCalls  int
		want       []models.UserListItem
		wantErr    error
	}{
		{
			name:       "returns the students",
			limit:      25,
			repoResult: students,
			wantCalls:  1,
			want:       students,
		},
		{
			name:    "rejects a limit above 100",
			limit:   101,
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:    "rejects a negative offset",
			limit:   25,
			offset:  -1,
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:      "wraps a repo error",
			limit:     25,
			repoErr:   errs.ErrDatabaseDown,
			wantCalls: 1,
			wantErr:   errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{studentsResult: tt.repoResult, studentsErr: tt.repoErr}
			svc := NewUserService(repo)

			got, err := svc.ListStudents(context.Background(), tt.limit, tt.offset, "moh")

			if repo.studentsCalls != tt.wantCalls {
				t.Fatalf("repo calls = %d, want %d", repo.studentsCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ListStudents: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserService_GetUserDetail(t *testing.T) {
	student := models.User{ID: 19, FirstName: "mohamad", LastName: "hassan", Number: 55}
	timeline := []models.UserActivityEvent{{Type: "visit", At: "2026-08-18T10:00:00Z", Summary: "visited home from phone", RefID: 4, DeviceType: "phone"}}

	tests := []struct {
		name           string
		idStr          string
		limit          int
		offset         int
		byIDErr        error
		activityErr    error
		lastDeviceErr  error
		lastDevice     string
		wantByID       int
		wantActivity   int
		wantLastDevice int
		want           models.UserDetail
		wantErr        error
	}{
		{
			name:           "returns the profile and its timeline",
			idStr:          " 19 ",
			limit:          10,
			lastDevice:     "phone",
			wantByID:       1,
			wantActivity:   1,
			wantLastDevice: 1,
			want:           models.UserDetail{User: student, LastDeviceType: "phone", Timeline: timeline},
		},
		{
			name:    "rejects a non numeric id",
			idStr:   "abc",
			limit:   10,
			wantErr: errs.ErrUserInvalidID,
		},
		{
			name:    "rejects an invalid limit",
			idStr:   "19",
			limit:   0,
			wantErr: errs.ErrInvalidParams,
		},
		{
			name:     "propagates a missing student",
			idStr:    "19",
			limit:    10,
			byIDErr:  errs.ErrUserNotFound,
			wantByID: 1,
			wantErr:  errs.ErrUserNotFound,
		},
		{
			name:         "wraps a timeline error",
			idStr:        "19",
			limit:        10,
			activityErr:  errs.ErrDatabaseDown,
			wantByID:     1,
			wantActivity: 1,
			wantErr:      errs.ErrDatabaseDown,
		},
		{
			name:           "wraps a last device error",
			idStr:          "19",
			limit:          10,
			lastDeviceErr:  errs.ErrDatabaseDown,
			wantByID:       1,
			wantActivity:   1,
			wantLastDevice: 1,
			wantErr:        errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepo{
				byIDResult:       student,
				byIDErr:          tt.byIDErr,
				activityResult:   timeline,
				activityErr:      tt.activityErr,
				lastDeviceResult: tt.lastDevice,
				lastDeviceErr:    tt.lastDeviceErr,
			}
			svc := NewUserService(repo)

			got, err := svc.GetUserDetail(context.Background(), tt.idStr, tt.limit, tt.offset)

			if repo.byIDCalls != tt.wantByID {
				t.Fatalf("get by id calls = %d, want %d", repo.byIDCalls, tt.wantByID)
			}
			if repo.activityCalls != tt.wantActivity {
				t.Fatalf("list activity calls = %d, want %d", repo.activityCalls, tt.wantActivity)
			}
			if repo.lastDeviceCalls != tt.wantLastDevice {
				t.Fatalf("last device calls = %d, want %d", repo.lastDeviceCalls, tt.wantLastDevice)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUserDetail: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserService_DeleteStaleGuests(t *testing.T) {
	t.Run("uses default ttl and returns count", func(t *testing.T) {
		repo := &fakeUserRepo{staleResult: 7}
		svc := NewUserService(repo)

		before := time.Now()
		got, err := svc.DeleteStaleGuests(context.Background(), 0)
		after := time.Now()
		if err != nil {
			t.Fatalf("DeleteStaleGuests: %v", err)
		}
		if got != 7 {
			t.Fatalf("deleted = %d, want 7", got)
		}
		if repo.staleCalls != 1 {
			t.Fatalf("repo calls = %d, want 1", repo.staleCalls)
		}
		wantEarliest := before.Add(-StaleGuestTTL)
		wantLatest := after.Add(-StaleGuestTTL)
		if repo.staleCutoff.Before(wantEarliest) || repo.staleCutoff.After(wantLatest) {
			t.Fatalf("cutoff %v outside [%v, %v]", repo.staleCutoff, wantEarliest, wantLatest)
		}
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repo := &fakeUserRepo{staleErr: errs.ErrDatabaseDown}
		svc := NewUserService(repo)

		_, err := svc.DeleteStaleGuests(context.Background(), StaleGuestTTL)
		if !errors.Is(err, errs.ErrDatabaseDown) {
			t.Fatalf("got %v, want %v", err, errs.ErrDatabaseDown)
		}
	})
}
