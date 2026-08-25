package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	uniqueViolationCode     = "23505"
	foreignKeyViolationCode = "23503"
)

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateGuest(ctx context.Context) (int, error) {
	var id int
	if err := r.db.QueryRowContext(ctx, insertNewGuestQuery).Scan(&id); err != nil {
		return 0, fmt.Errorf("create guest: %w", err)
	}
	return id, nil
}

func (r *postgresUserRepository) ClaimGuest(ctx context.Context, guestID int, u models.User) (models.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, claimGuestQuery, u.FirstName, u.LastName, u.Number, guestID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrUserGuestNotFound
		}
		if isUniqueViolation(err) {
			return models.User{}, errs.ErrUsernameTaken
		}
		return models.User{}, fmt.Errorf("claim guest: %w", err)
	}
	return user, nil
}

// AdoptGuest moves a guest's activity onto an existing student, then deletes
// the guest row. Callers must reassign before delete: page_views cascade.
func (r *postgresUserRepository) AdoptGuest(ctx context.Context, guestID int, userID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin adopt guest: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var locked int
	if err := tx.QueryRowContext(ctx, lockGuestForAdoptQuery, guestID).Scan(&locked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrUserGuestNotFound
		}
		return fmt.Errorf("lock guest: %w", err)
	}

	for _, q := range []string{
		reassignPageViewsQuery,
		reassignLinkClicksQuery,
		reassignReportsQuery,
		reassignContributionsQuery,
		reassignFeedbackQuery,
		reassignFavoriteEventsQuery,
	} {
		if _, err := tx.ExecContext(ctx, q, guestID, userID); err != nil {
			return fmt.Errorf("reassign guest activity: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, deleteGuestQuery, guestID); err != nil {
		return fmt.Errorf("delete guest: %w", err)
	}
	if _, err := tx.ExecContext(ctx, touchLastSeenQuery, userID); err != nil {
		return fmt.Errorf("touch last seen: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit adopt guest: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, u models.User) (models.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, insertNewUserQuery, u.FirstName, u.LastName, u.Number))
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, errs.ErrUsernameTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int) (models.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, getUserByIDQuery, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) GetByCredentials(ctx context.Context, u models.User) (models.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, getUserByCredentialsQuery, u.FirstName, u.LastName, u.Number))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by credentials: %w", err)
	}
	return user, nil
}

func (r *postgresUserRepository) AddFavorite(ctx context.Context, userID int, courseID int) error {
	return r.toggleFavorite(ctx, userID, courseID, addFavoriteQuery, "added")
}

func (r *postgresUserRepository) RemoveFavorite(ctx context.Context, userID int, courseID int) error {
	return r.toggleFavorite(ctx, userID, courseID, removeFavoriteQuery, "removed")
}

// toggleFavorite keeps the live favorites array and the event history in sync by
// writing both in one transaction. A toggle that changes nothing writes no event.
func (r *postgresUserRepository) toggleFavorite(ctx context.Context, userID int, courseID int, updateQuery string, action string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin favorite tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, updateQuery, userID, courseID)
	if err != nil {
		return fmt.Errorf("update favorite course ids: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update favorite course ids rows affected: %w", err)
	}
	if affected == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, insertFavoriteEventQuery, userID, courseID, action); err != nil {
		if isForeignKeyViolation(err) {
			return errs.ErrCourseNotFound
		}
		return fmt.Errorf("insert favorite event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit favorite tx: %w", err)
	}
	return nil
}

// ── Admin Queries ───────────────────────────────────────────────────────────

func (r *postgresUserRepository) ListStudents(ctx context.Context, limit int, offset int, q string) ([]models.UserListItem, error) {
	query, args := listStudentsQuery, []any{limit, offset}
	if q != "" {
		query, args = listStudentsWithQQuery, []any{"%" + q + "%", limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list students query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	students := []models.UserListItem{}
	for rows.Next() {
		var s models.UserListItem
		if err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Number, &s.CreatedAt, &s.LastSeenAt, &s.VisitCount, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("list students rows scan: %w", err)
		}
		s.Handle = models.UserHandle(s.FirstName, s.LastName, s.Number, s.ID)
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list students rows err: %w", err)
	}
	return students, nil
}

func (r *postgresUserRepository) ListActivity(ctx context.Context, userID int, limit int, offset int) ([]models.UserActivityEvent, error) {
	rows, err := r.db.QueryContext(ctx, listUserTimelineQuery, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list user activity query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	events := []models.UserActivityEvent{}
	for rows.Next() {
		var e models.UserActivityEvent
		if err := rows.Scan(&e.Type, &e.At, &e.Summary, &e.RefID, &e.DeviceType); err != nil {
			return nil, fmt.Errorf("list user activity rows scan: %w", err)
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user activity rows err: %w", err)
	}
	return events, nil
}

func (r *postgresUserRepository) GetLastDeviceType(ctx context.Context, userID int) (string, error) {
	var deviceType string
	err := r.db.QueryRowContext(ctx, getLastDeviceTypeQuery, userID).Scan(&deviceType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get last device type: %w", err)
	}
	return deviceType, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (models.User, error) {
	var u models.User
	var favoriteIDs string
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Number, &u.IsGuest, &favoriteIDs, &u.CreatedAt, &u.LastSeenAt, &u.PreferedLang, &u.PreferedTheme); err != nil {
		return models.User{}, err
	}

	u.FavoriteCourseIDs = []int{}
	if err := json.Unmarshal([]byte(favoriteIDs), &u.FavoriteCourseIDs); err != nil {
		return models.User{}, fmt.Errorf("decode favorite course ids: %w", err)
	}
	if u.FavoriteCourseIDs == nil {
		u.FavoriteCourseIDs = []int{}
	}
	u.Handle = models.UserHandle(u.FirstName, u.LastName, u.Number, u.ID)

	return u, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationCode
}
