package api

import (
	"errors"
	"net/http"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/models"
)

// credentialsBody is the signup and login payload. Identity flags never come from
// the client: is_guest and the user id are decided by the server.
type credentialsBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Number    int    `json:"number"`
}

func (h *Handler) handlePostGuest(w http.ResponseWriter, r *http.Request) {
	userID, err := h.userService.CreateGuest(r.Context())
	if err != nil {
		h.LoggerWithID(r).Error("create guest failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	token, err := generateUserToken(userID, true, h.jwtSecret)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Failed to create guest jwt")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *Handler) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var body credentialsBody
	if !decodeJSONBody(w, r, &body) {
		return
	}

	// A guest token is optional: when it is valid the guest row is claimed so the
	// pre-signup history stays attached to the same student.
	guestID := middleware.GuestIDFromHeader(string(h.jwtSecret), r.Header.Get("Authorization"))

	user, err := h.userService.RegisterUser(r.Context(), guestID, bodyToUser(body))
	if err != nil {
		mapRegisterUserErr(h, w, r, err)
		return
	}

	token, err := generateUserToken(user.ID, false, h.jwtSecret)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Failed to create user jwt")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

func (h *Handler) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	var body credentialsBody
	if !decodeJSONBody(w, r, &body) {
		return
	}

	// Same optional guest token as register: the first visit this browser
	// recorded as a guest is reassigned onto the student we just found.
	guestID := middleware.GuestIDFromHeader(string(h.jwtSecret), r.Header.Get("Authorization"))

	user, err := h.userService.LoginUser(r.Context(), guestID, bodyToUser(body))
	if err != nil {
		mapLoginUserErr(h, w, r, err)
		return
	}

	token, err := generateUserToken(user.ID, false, h.jwtSecret)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Failed to create user jwt")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		mapGetUserErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleAddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	if err := h.userService.AddFavorite(r.Context(), userID, r.PathValue("course_id")); err != nil {
		mapFavoriteErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	if err := h.userService.RemoveFavorite(r.Context(), userID, r.PathValue("course_id")); err != nil {
		mapFavoriteErr(h, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Admin Protected Handlers ────────────────────────────────────────────────

func (h *Handler) handleAdminGetUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset, q := parsePaginationParams(r, 25)

	students, err := h.userService.ListStudents(r.Context(), limit, offset, q)
	if err != nil {
		mapListStudentsErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, students)
}

func (h *Handler) handleAdminGetUser(w http.ResponseWriter, r *http.Request) {
	limit, offset, _ := parsePaginationParams(r, 25)

	detail, err := h.userService.GetUserDetail(r.Context(), r.PathValue("id"), limit, offset)
	if err != nil {
		mapUserDetailErr(h, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// Helpers functions

func bodyToUser(body credentialsBody) models.User {
	return models.User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Number:    body.Number,
	}
}

func mapRegisterUserErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrUsernameTaken):
		writeJSONError(w, r, http.StatusConflict, "This name and number are already taken, try another number")
	case errors.Is(err, errs.ErrUserNameRequired):
		writeJSONError(w, r, http.StatusBadRequest, "First and last name are required")
	case errors.Is(err, errs.ErrUserNumberRange):
		writeJSONError(w, r, http.StatusBadRequest, "Number must be between 1 and 100")
	default:
		h.LoggerWithID(r).Error("register user failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapLoginUserErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		writeJSONError(w, r, http.StatusNotFound, "No student found with this name and number, sign up first")
	case errors.Is(err, errs.ErrUserNameRequired):
		writeJSONError(w, r, http.StatusBadRequest, "First and last name are required")
	case errors.Is(err, errs.ErrUserNumberRange):
		writeJSONError(w, r, http.StatusBadRequest, "Number must be between 1 and 100")
	default:
		h.LoggerWithID(r).Error("login user failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapGetUserErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		writeJSONError(w, r, http.StatusNotFound, "User not found")
	case errors.Is(err, errs.ErrUserInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid user id")
	default:
		h.LoggerWithID(r).Error("get user failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapFavoriteErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrCourseInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid course id")
	case errors.Is(err, errs.ErrCourseNotFound):
		writeJSONError(w, r, http.StatusNotFound, "Course not found")
	default:
		h.LoggerWithID(r).Error("update favorites failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapListStudentsErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, r, http.StatusBadRequest, "Limit should be between 1-100 and Offset >= 0")
	default:
		h.LoggerWithID(r).Error("list students failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func mapUserDetailErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrUserInvalidID):
		writeJSONError(w, r, http.StatusBadRequest, "Invalid user id")
	case errors.Is(err, errs.ErrUserNotFound):
		writeJSONError(w, r, http.StatusNotFound, "User not found")
	case errors.Is(err, errs.ErrInvalidParams):
		writeJSONError(w, r, http.StatusBadRequest, "Limit should be between 1-100 and Offset >= 0")
	default:
		h.LoggerWithID(r).Error("get user detail failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}
