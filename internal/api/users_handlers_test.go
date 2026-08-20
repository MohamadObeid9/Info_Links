package api

import (
	"context"
	"encoding/json"
	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeUserService struct {
	getCalls  int
	getResult int
	getErr    error

	registerCalls   int
	registerGuestID int
	registerUser    models.User
	registerResult  models.User
	registerErr     error

	loginGuestID int
	loginResult  models.User
	loginErr     error

	meResult models.User
	meErr    error

	favoriteCalls    int
	favoriteUserID   int
	favoriteCourseID string
	favoriteErr      error

	listResult []models.UserListItem
	listErr    error

	detailResult models.UserDetail
	detailErr    error
}

func (f *fakeUserService) CreateGuest(ctx context.Context) (int, error) {
	f.getCalls++
	if f.getErr != nil {
		return 0, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeUserService) RegisterUser(ctx context.Context, guestID int, u models.User) (models.User, error) {
	f.registerCalls++
	f.registerGuestID = guestID
	f.registerUser = u
	if f.registerErr != nil {
		return models.User{}, f.registerErr
	}
	return f.registerResult, nil
}

func (f *fakeUserService) LoginUser(ctx context.Context, guestID int, u models.User) (models.User, error) {
	f.loginGuestID = guestID
	if f.loginErr != nil {
		return models.User{}, f.loginErr
	}
	return f.loginResult, nil
}

func (f *fakeUserService) GetUser(ctx context.Context, userID int) (models.User, error) {
	if f.meErr != nil {
		return models.User{}, f.meErr
	}
	return f.meResult, nil
}

func (f *fakeUserService) AddFavorite(ctx context.Context, userID int, courseIDStr string) error {
	f.favoriteCalls++
	f.favoriteUserID = userID
	f.favoriteCourseID = courseIDStr
	return f.favoriteErr
}

func (f *fakeUserService) RemoveFavorite(ctx context.Context, userID int, courseIDStr string) error {
	f.favoriteCalls++
	f.favoriteUserID = userID
	f.favoriteCourseID = courseIDStr
	return f.favoriteErr
}

func (f *fakeUserService) ListStudents(ctx context.Context, limit int, offset int, q string) ([]models.UserListItem, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeUserService) GetUserDetail(ctx context.Context, idStr string, limit int, offset int) (models.UserDetail, error) {
	if f.detailErr != nil {
		return models.UserDetail{}, f.detailErr
	}
	return f.detailResult, nil
}

func TestHandleCreateGuest(t *testing.T) {
	userID := 5
	tests := []struct {
		name         string
		getResult    int
		getErr       error
		statusWanted int
		errMsg       string
		wantCalls    int
	}{
		{
			name:         "returns 201 status",
			getResult:    userID,
			statusWanted: http.StatusCreated,
			wantCalls:    1,
		},
		{
			name:         "500 when service fails",
			getErr:       errs.ErrDatabaseDown,
			statusWanted: http.StatusInternalServerError,
			errMsg:       "Internal server error",
			wantCalls:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{getResult: tt.getResult, getErr: tt.getErr}
			h := testHandler(t, withUser(fakeUser))
			req := httptest.NewRequest(http.MethodPost, "/api/users/guest", nil)
			rr := httptest.NewRecorder()

			h.handlePostGuest(rr, req)

			if fakeUser.getCalls != tt.wantCalls {
				t.Fatalf("get calls = %d, want %d", fakeUser.getCalls, tt.wantCalls)
			}
			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg != "" {
				var body map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body["error"] != tt.errMsg {
					t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
				}
				return
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
			var body map[string]any
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if token, _ := body["token"].(string); token == "" {
				t.Fatalf("body = %v, want a non-empty token field", body)
			}
		})
	}
}

func TestHandleRegisterUser(t *testing.T) {
	registered := models.User{ID: 7, FirstName: "mohamad", LastName: "hassan", Number: 55, Handle: "mohamad_hassan_55"}

	tests := []struct {
		name         string
		body         string
		registerErr  error
		statusWanted int
		errMsg       string
	}{
		{
			name:         "201 with token and user",
			body:         `{"first_name":"mohamad","last_name":"hassan","number":55}`,
			statusWanted: http.StatusCreated,
		},
		{
			name:         "409 when the name and number are taken",
			body:         `{"first_name":"mohamad","last_name":"hassan","number":55}`,
			registerErr:  errs.ErrUsernameTaken,
			statusWanted: http.StatusConflict,
			errMsg:       "This name and number are already taken, try another number",
		},
		{
			name:         "400 when the number is out of range",
			body:         `{"first_name":"mohamad","last_name":"hassan","number":101}`,
			registerErr:  errs.ErrUserNumberRange,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Number must be between 1 and 100",
		},
		{
			name:         "400 on a malformed body",
			body:         `{"first_name":`,
			statusWanted: http.StatusBadRequest,
			errMsg:       "Invalid request body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{registerResult: registered, registerErr: tt.registerErr}
			h := testHandler(t, withUser(fakeUser))
			req := jsonRequest(http.MethodPost, "/api/users/register", tt.body)
			rr := httptest.NewRecorder()

			h.handleRegisterUser(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg != "" {
				var body map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("decode error body: %v", err)
				}
				if body["error"] != tt.errMsg {
					t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
				}
				return
			}
			var body struct {
				Token string      `json:"token"`
				User  models.User `json:"user"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Token == "" {
				t.Fatal("token is empty")
			}
			if body.User.Handle != registered.Handle {
				t.Fatalf("handle = %q, want %q", body.User.Handle, registered.Handle)
			}
		})
	}
}

// A valid guest token must be claimed, so pre-signup history stays on the same row.
func TestHandleRegisterUserClaimsGuestFromToken(t *testing.T) {
	const guestID = 31

	guestToken, err := generateUserToken(guestID, true, []byte("test-jwt-secret"))
	if err != nil {
		t.Fatalf("generateUserToken: %v", err)
	}

	tests := []struct {
		name        string
		authHeader  string
		wantGuestID int
	}{
		{name: "claims the guest id from a guest token", authHeader: "Bearer " + guestToken, wantGuestID: guestID},
		{name: "no header means a fresh student", authHeader: "", wantGuestID: 0},
		{name: "garbage header is ignored", authHeader: "Bearer not-a-token", wantGuestID: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{registerResult: models.User{ID: 7}}
			h := testHandler(t, withUser(fakeUser))
			req := jsonRequest(http.MethodPost, "/api/users/register", `{"first_name":"mo","last_name":"h","number":5}`)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			h.handleRegisterUser(rr, req)

			if rr.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
			}
			if fakeUser.registerGuestID != tt.wantGuestID {
				t.Fatalf("guest id = %d, want %d", fakeUser.registerGuestID, tt.wantGuestID)
			}
		})
	}
}

func TestHandleLoginUser(t *testing.T) {
	tests := []struct {
		name         string
		loginErr     error
		statusWanted int
		errMsg       string
	}{
		{
			name:         "200 with token and user",
			statusWanted: http.StatusOK,
		},
		{
			name:         "404 when no student matches",
			loginErr:     errs.ErrUserNotFound,
			statusWanted: http.StatusNotFound,
			errMsg:       "No student found with this name and number, sign up first",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{loginResult: models.User{ID: 7, Handle: "mo_h_5"}, loginErr: tt.loginErr}
			h := testHandler(t, withUser(fakeUser))
			req := jsonRequest(http.MethodPost, "/api/users/login", `{"first_name":"mo","last_name":"h","number":5}`)
			rr := httptest.NewRecorder()

			h.handleLoginUser(rr, req)

			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if tt.errMsg == "" {
				return
			}
			var body map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("decode error body: %v", err)
			}
			if body["error"] != tt.errMsg {
				t.Fatalf("error = %q, want %q", body["error"], tt.errMsg)
			}
		})
	}
}

func TestHandleLoginUserAdoptsGuestFromToken(t *testing.T) {
	const guestID = 31

	guestToken, err := generateUserToken(guestID, true, []byte("test-jwt-secret"))
	if err != nil {
		t.Fatalf("generateUserToken: %v", err)
	}
	registeredToken, err := generateUserToken(7, false, []byte("test-jwt-secret"))
	if err != nil {
		t.Fatalf("generateUserToken: %v", err)
	}

	tests := []struct {
		name        string
		authHeader  string
		wantGuestID int
	}{
		{name: "adopts the guest id from a guest token", authHeader: "Bearer " + guestToken, wantGuestID: guestID},
		{name: "no header means nothing to adopt", authHeader: "", wantGuestID: 0},
		{name: "a registered token is not treated as a guest", authHeader: "Bearer " + registeredToken, wantGuestID: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{loginResult: models.User{ID: 7, Handle: "mo_h_5"}}
			h := testHandler(t, withUser(fakeUser))
			req := jsonRequest(http.MethodPost, "/api/users/login", `{"first_name":"mo","last_name":"h","number":5}`)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			h.handleLoginUser(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if fakeUser.loginGuestID != tt.wantGuestID {
				t.Fatalf("guest id = %d, want %d", fakeUser.loginGuestID, tt.wantGuestID)
			}
		})
	}
}

func TestHandleGetMe(t *testing.T) {
	t.Run("200 for an authenticated student", func(t *testing.T) {
		fakeUser := &fakeUserService{meResult: models.User{ID: testStudentID, Handle: "mo_h_5"}}
		h := testHandler(t, withUser(fakeUser))
		rr := httptest.NewRecorder()

		h.handleGetMe(rr, studentRequest(http.MethodGet, "/api/users/me", ""))

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("401 without a student in context", func(t *testing.T) {
		h := testHandler(t, withUser(&fakeUserService{}))
		rr := httptest.NewRecorder()

		h.handleGetMe(rr, jsonRequest(http.MethodGet, "/api/users/me", ""))

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}

// The favorite owner comes from the token context, never from the request.
func TestHandleFavorites(t *testing.T) {
	tests := []struct {
		name         string
		remove       bool
		favoriteErr  error
		statusWanted int
	}{
		{name: "204 on add", statusWanted: http.StatusNoContent},
		{name: "204 on remove", remove: true, statusWanted: http.StatusNoContent},
		{name: "400 on an invalid course id", favoriteErr: errs.ErrCourseInvalidID, statusWanted: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := &fakeUserService{favoriteErr: tt.favoriteErr}
			h := testHandler(t, withUser(fakeUser))
			req := studentRequest(http.MethodPost, "/api/users/me/favorites/12", "")
			req.SetPathValue("course_id", "12")
			rr := httptest.NewRecorder()

			if tt.remove {
				h.handleRemoveFavorite(rr, req)
			} else {
				h.handleAddFavorite(rr, req)
			}

			if rr.Code != tt.statusWanted {
				t.Fatalf("status = %d, want %d", rr.Code, tt.statusWanted)
			}
			if fakeUser.favoriteUserID != testStudentID {
				t.Fatalf("user id = %d, want %d from the token context", fakeUser.favoriteUserID, testStudentID)
			}
			if fakeUser.favoriteCourseID != "12" {
				t.Fatalf("course id = %q, want %q", fakeUser.favoriteCourseID, "12")
			}
		})
	}
}
