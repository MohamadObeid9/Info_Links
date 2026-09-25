package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
	"infolinks-backend/internal/repository"
)

// StaleGuestTTL is how long an unclaimed guest may sit idle before cleanup.
const StaleGuestTTL = 24 * time.Hour

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateGuest(ctx context.Context) (int, error) {
	id, err := s.repo.CreateGuest(ctx)
	if err != nil {
		return 0, fmt.Errorf("create guest: %w", err)
	}
	return id, nil
}

// DeleteStaleGuests drops unclaimed guests idle longer than ttl (defaults to StaleGuestTTL).
func (s *UserService) DeleteStaleGuests(ctx context.Context, ttl time.Duration) (int64, error) {
	if ttl <= 0 {
		ttl = StaleGuestTTL
	}
	n, err := s.repo.DeleteStaleGuests(ctx, time.Now().Add(-ttl))
	if err != nil {
		return 0, fmt.Errorf("delete stale guests: %w", err)
	}
	return n, nil
}

// RegisterUser claims the guest row identified by guestID so pre-signup activity
// keeps the same user id. A guest id that no longer exists (stale token, already
// claimed) falls through to a brand new student.
func (s *UserService) RegisterUser(ctx context.Context, guestID int, u models.User) (models.User, error) {
	u, err := normalizeCredentials(u)
	if err != nil {
		return models.User{}, err
	}

	if guestID > 0 {
		claimed, err := s.repo.ClaimGuest(ctx, guestID, u)
		switch {
		case err == nil:
			return claimed, nil
		case errors.Is(err, errs.ErrUserGuestNotFound):
		default:
			return models.User{}, fmt.Errorf("claim guest: %w", err)
		}
	}

	created, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return models.User{}, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

func (s *UserService) LoginUser(ctx context.Context, guestID int, u models.User) (models.User, error) {
	u, err := normalizeCredentials(u)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.repo.GetByCredentials(ctx, u)
	if err != nil {
		return models.User{}, fmt.Errorf("login user: %w", err)
	}

	// The first visit this browser recorded still belongs to the guest row.
	// sessionStorage will not send a second page view after sign-in, so those
	// rows have to move or they stay labelled guest_<id> forever.
	if guestID > 0 && guestID != user.ID {
		if err := s.repo.AdoptGuest(ctx, guestID, user.ID); err != nil && !errors.Is(err, errs.ErrUserGuestNotFound) {
			return models.User{}, fmt.Errorf("adopt guest: %w", err)
		}
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, userID int) (models.User, error) {
	if userID <= 0 {
		return models.User{}, errs.ErrUserInvalidID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) AddFavorite(ctx context.Context, userID int, courseIDStr string) error {
	courseID, err := parseCourseID(courseIDStr)
	if err != nil {
		return err
	}
	if err := s.repo.AddFavorite(ctx, userID, courseID); err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (s *UserService) RemoveFavorite(ctx context.Context, userID int, courseIDStr string) error {
	courseID, err := parseCourseID(courseIDStr)
	if err != nil {
		return err
	}
	if err := s.repo.RemoveFavorite(ctx, userID, courseID); err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

func (s *UserService) ListStudents(ctx context.Context, limit int, offset int, q string, sort string, order string) ([]models.UserListItem, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, errs.ErrInvalidParams
	}

	params, err := normalizeStudentListParams(limit, offset, q, sort, order)
	if err != nil {
		return nil, err
	}

	students, err := s.repo.ListStudents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	return students, nil
}

// DeleteStudents removes registered students and their cascaded analytics.
// ids must be non-empty, positive, and at most 100 unique values.
func (s *UserService) DeleteStudents(ctx context.Context, ids []int) (int64, error) {
	cleaned, err := normalizeStudentIDs(ids)
	if err != nil {
		return 0, err
	}

	n, err := s.repo.DeleteStudents(ctx, cleaned)
	if err != nil {
		return 0, fmt.Errorf("delete students: %w", err)
	}
	if n == 0 {
		return 0, errs.ErrUserNotFound
	}
	return n, nil
}

func (s *UserService) GetUserDetail(ctx context.Context, idStr string, limit int, offset int) (models.UserDetail, error) {
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil || id <= 0 {
		return models.UserDetail{}, errs.ErrUserInvalidID
	}
	if limit <= 0 || limit > 100 || offset < 0 {
		return models.UserDetail{}, errs.ErrInvalidParams
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.UserDetail{}, fmt.Errorf("get user detail: %w", err)
	}

	timeline, err := s.repo.ListActivity(ctx, id, limit, offset)
	if err != nil {
		return models.UserDetail{}, fmt.Errorf("list user activity: %w", err)
	}

	lastDevice, err := s.repo.GetLastDeviceType(ctx, id)
	if err != nil {
		return models.UserDetail{}, fmt.Errorf("get last device type: %w", err)
	}

	return models.UserDetail{User: user, LastDeviceType: lastDevice, Timeline: timeline}, nil
}

// normalizeCredentials trims and lowercases the name so the partial unique index
// treats "Ali" and "ali" as the same student.
func normalizeCredentials(u models.User) (models.User, error) {
	u.FirstName = strings.ToLower(strings.TrimSpace(u.FirstName))
	u.LastName = strings.ToLower(strings.TrimSpace(u.LastName))
	if u.FirstName == "" || u.LastName == "" {
		return models.User{}, errs.ErrUserNameRequired
	}

	if u.Number < 1 || u.Number > 100 {
		return models.User{}, errs.ErrUserNumberRange
	}
	return u, nil
}

func normalizeStudentIDs(ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, errs.ErrUserInvalidID
	}
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, errs.ErrUserInvalidID
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, errs.ErrUserInvalidID
	}
	if len(out) > 100 {
		return nil, errs.ErrInvalidParams
	}
	return out, nil
}

func normalizeStudentListParams(limit, offset int, q, sort, order string) (repository.StudentListParams, error) {
	sort = strings.ToLower(strings.TrimSpace(sort))
	order = strings.ToLower(strings.TrimSpace(order))
	if sort == "" {
		sort = "name"
	}
	if order == "" {
		order = "asc"
	}

	switch sort {
	case "name", "first_seen", "last_seen", "visits", "clicks", "favorites":
	default:
		return repository.StudentListParams{}, errs.ErrUserInvalidSort
	}
	switch order {
	case "asc", "desc":
	default:
		return repository.StudentListParams{}, errs.ErrUserInvalidOrder
	}

	return repository.StudentListParams{
		Limit:  limit,
		Offset: offset,
		Q:      strings.TrimSpace(q),
		Sort:   sort,
		Order:  order,
	}, nil
}

func parseCourseID(courseIDStr string) (int, error) {
	courseID, err := strconv.Atoi(strings.TrimSpace(courseIDStr))
	if err != nil || courseID <= 0 {
		return 0, errs.ErrCourseInvalidID
	}
	return courseID, nil
}
