package repository

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func newTestUserRepo(t *testing.T) (UserRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPostgresUserRepository(db), mock
}

var userColumnNames = []string{"id", "first_name", "last_name", "number", "is_guest", "favorite_course_ids", "created_at", "last_seen_at", "prefered_lang", "prefered_theme"}

// userRow builds one users row as the queries return it, favorites as JSON text.
func userRow(id int, firstName, lastName string, number int, isGuest bool, favorites string) *sqlmock.Rows {
	return sqlmock.NewRows(userColumnNames).
		AddRow(id, firstName, lastName, number, isGuest, favorites, "2026-08-18T10:00:00Z", "2026-08-18T11:00:00Z", "eng", "system")
}

func wantUser(id int, firstName, lastName string, number int, isGuest bool, favorites []int) models.User {
	return models.User{
		ID:                id,
		FirstName:         firstName,
		LastName:          lastName,
		Number:            number,
		IsGuest:           isGuest,
		Handle:            models.UserHandle(firstName, lastName, number, id),
		FavoriteCourseIDs: favorites,
		CreatedAt:         "2026-08-18T10:00:00Z",
		LastSeenAt:        "2026-08-18T11:00:00Z",
		PreferedLang:      "eng",
		PreferedTheme:     "system",
	}
}

func uniqueViolation() error {
	return &pgconn.PgError{Code: uniqueViolationCode, Message: "duplicate key value violates unique constraint"}
}

func TestUserRepository_CreateGuest(t *testing.T) {
	columns := []string{"id"}

	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     int
		err      error
	}{
		{
			name:     "query error",
			want:     0,
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
		{
			name: "returns id",
			rows: sqlmock.NewRows(columns).AddRow(19),
			want: 19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			exp := mock.ExpectQuery(insertNewGuestQuery)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.CreateGuest(context.Background())
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_ClaimGuest(t *testing.T) {
	tests := []struct {
		name     string
		guestID  int
		queryErr error
		rows     *sqlmock.Rows
		want     models.User
		err      error
	}{
		{
			name:    "claim keeps the guest id",
			guestID: 19,
			rows:    userRow(19, "mohamad", "hassan", 55, false, "[1,2]"),
			want:    wantUser(19, "mohamad", "hassan", 55, false, []int{1, 2}),
		},
		{
			name:     "stale guest id",
			guestID:  404,
			queryErr: sql.ErrNoRows,
			err:      errs.ErrUserGuestNotFound,
		},
		{
			name:     "name already taken",
			guestID:  19,
			queryErr: uniqueViolation(),
			err:      errs.ErrUsernameTaken,
		},
		{
			name:     "query error",
			guestID:  19,
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)
			user := models.User{FirstName: "mohamad", LastName: "hassan", Number: 55}

			exp := mock.ExpectQuery(claimGuestQuery).WithArgs(user.FirstName, user.LastName, user.Number, tt.guestID)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.ClaimGuest(context.Background(), tt.guestID, user)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_AdoptGuest(t *testing.T) {
	const guestID, userID = 42, 19

	reassignQueries := []string{
		reassignPageViewsQuery,
		reassignLinkClicksQuery,
		reassignReportsQuery,
		reassignContributionsQuery,
		reassignFeedbackQuery,
		reassignFavoriteEventsQuery,
		reassignSearchEventsQuery,
		reassignBrowseEventsQuery,
	}

	t.Run("reassigns activity then deletes the guest", func(t *testing.T) {
		repo, mock := newTestUserRepo(t)

		mock.ExpectBegin()
		mock.ExpectQuery(lockGuestForAdoptQuery).WithArgs(guestID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(guestID))
		for _, q := range reassignQueries {
			mock.ExpectExec(q).WithArgs(guestID, userID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}
		mock.ExpectExec(deleteGuestQuery).WithArgs(guestID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(touchLastSeenQuery).WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.AdoptGuest(context.Background(), guestID, userID)
		assertRepoErr(t, mock, err, nil)
	})

	t.Run("stale guest id", func(t *testing.T) {
		repo, mock := newTestUserRepo(t)

		mock.ExpectBegin()
		mock.ExpectQuery(lockGuestForAdoptQuery).WithArgs(guestID).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		err := repo.AdoptGuest(context.Background(), guestID, userID)
		assertRepoErr(t, mock, err, errs.ErrUserGuestNotFound)
	})

	t.Run("lock error", func(t *testing.T) {
		repo, mock := newTestUserRepo(t)

		mock.ExpectBegin()
		mock.ExpectQuery(lockGuestForAdoptQuery).WithArgs(guestID).
			WillReturnError(errs.ErrDatabaseDown)
		mock.ExpectRollback()

		err := repo.AdoptGuest(context.Background(), guestID, userID)
		assertRepoErr(t, mock, err, errs.ErrDatabaseDown)
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     models.User
		err      error
	}{
		{
			name: "creates a student with empty favorites",
			rows: userRow(7, "mohamad", "hassan", 55, false, "[]"),
			want: wantUser(7, "mohamad", "hassan", 55, false, []int{}),
		},
		{
			name:     "name and number already taken",
			queryErr: uniqueViolation(),
			err:      errs.ErrUsernameTaken,
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)
			user := models.User{FirstName: "mohamad", LastName: "hassan", Number: 55}

			exp := mock.ExpectQuery(insertNewUserQuery).WithArgs(user.FirstName, user.LastName, user.Number)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.CreateUser(context.Background(), user)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     models.User
		err      error
	}{
		{
			name: "returns the student",
			rows: userRow(7, "mohamad", "hassan", 55, false, "[3]"),
			want: wantUser(7, "mohamad", "hassan", 55, false, []int{3}),
		},
		{
			name: "guest falls back to a guest handle",
			rows: userRow(7, "", "", 0, true, "[]"),
			want: wantUser(7, "", "", 0, true, []int{}),
		},
		{
			name:     "user not found",
			queryErr: sql.ErrNoRows,
			err:      errs.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			exp := mock.ExpectQuery(getUserByIDQuery).WithArgs(7)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.GetByID(context.Background(), 7)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_GetByCredentials(t *testing.T) {
	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     models.User
		err      error
	}{
		{
			name: "returns the student",
			rows: userRow(7, "mohamad", "hassan", 55, false, "[]"),
			want: wantUser(7, "mohamad", "hassan", 55, false, []int{}),
		},
		{
			name:     "no student with this name and number",
			queryErr: sql.ErrNoRows,
			err:      errs.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)
			user := models.User{FirstName: "mohamad", LastName: "hassan", Number: 55}

			exp := mock.ExpectQuery(getUserByCredentialsQuery).WithArgs(user.FirstName, user.LastName, user.Number)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.GetByCredentials(context.Background(), user)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_Favorites(t *testing.T) {
	const (
		userID   = 7
		courseID = 3
	)

	tests := []struct {
		name        string
		remove      bool
		affected    int64
		updateErr   error
		eventErr    error
		wantEvent   bool
		wantCommit  bool
		wantRollbck bool
		err         error
	}{
		{
			name:       "add writes the array and the event in one tx",
			affected:   1,
			wantEvent:  true,
			wantCommit: true,
		},
		{
			name:       "remove writes the array and the event in one tx",
			remove:     true,
			affected:   1,
			wantEvent:  true,
			wantCommit: true,
		},
		{
			name:        "adding an already favorited course writes no event",
			affected:    0,
			wantRollbck: true,
		},
		{
			name:        "removing a course that is not a favorite writes no event",
			remove:      true,
			affected:    0,
			wantRollbck: true,
		},
		{
			name:        "unknown course id",
			affected:    1,
			wantEvent:   true,
			eventErr:    &pgconn.PgError{Code: foreignKeyViolationCode},
			wantRollbck: true,
			err:         errs.ErrCourseNotFound,
		},
		{
			name:        "array update error",
			updateErr:   errs.ErrDatabaseDown,
			wantRollbck: true,
			err:         errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			updateQuery, action := addFavoriteQuery, "added"
			if tt.remove {
				updateQuery, action = removeFavoriteQuery, "removed"
			}

			mock.ExpectBegin()
			update := mock.ExpectExec(updateQuery).WithArgs(userID, courseID)
			if tt.updateErr != nil {
				update.WillReturnError(tt.updateErr)
			} else {
				update.WillReturnResult(sqlmock.NewResult(0, tt.affected))
			}
			if tt.wantEvent {
				event := mock.ExpectExec(insertFavoriteEventQuery).WithArgs(userID, courseID, action)
				if tt.eventErr != nil {
					event.WillReturnError(tt.eventErr)
				} else {
					event.WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}
			if tt.wantCommit {
				mock.ExpectCommit()
			}
			if tt.wantRollbck {
				mock.ExpectRollback()
			}

			var err error
			if tt.remove {
				err = repo.RemoveFavorite(context.Background(), userID, courseID)
			} else {
				err = repo.AddFavorite(context.Background(), userID, courseID)
			}
			assertRepoErr(t, mock, err, tt.err)
		})
	}
}

func TestUserRepository_ListStudents(t *testing.T) {
	columns := []string{"id", "first_name", "last_name", "number", "created_at", "last_seen_at", "visit_count", "click_count"}

	tests := []struct {
		name      string
		q         string
		query     string
		queryArgs []any
		queryErr  error
		rows      [][]any
		want      []models.UserListItem
		err       error
	}{
		{
			name:      "list without search",
			query:     listStudentsQuery,
			queryArgs: []any{25, 0},
			rows: [][]any{
				{1, "mohamad", "hassan", 55, "2026-08-01T10:00:00Z", "2026-08-18T10:00:00Z", 12, 4},
			},
			want: []models.UserListItem{
				{
					ID: 1, Handle: "mohamad_hassan_55", FirstName: "mohamad", LastName: "hassan", Number: 55,
					CreatedAt: "2026-08-01T10:00:00Z", LastSeenAt: "2026-08-18T10:00:00Z", VisitCount: 12, ClickCount: 4,
				},
			},
		},
		{
			name:      "list with search",
			q:         "moh",
			query:     listStudentsWithQQuery,
			queryArgs: []any{"%moh%", 10, 5},
			rows:      [][]any{},
			want:      []models.UserListItem{},
		},
		{
			name:      "query error",
			query:     listStudentsQuery,
			queryArgs: []any{25, 0},
			queryErr:  errs.ErrDatabaseDown,
			err:       errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			limit, offset := 25, 0
			if tt.q != "" {
				limit, offset = 10, 5
			}

			exp := mock.ExpectQuery(tt.query).WithArgs(driverValues(tt.queryArgs)...)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				rows := sqlmock.NewRows(columns)
				for _, row := range tt.rows {
					rows.AddRow(driverValues(row)...)
				}
				exp.WillReturnRows(rows)
			}

			got, err := repo.ListStudents(context.Background(), limit, offset, tt.q)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_ListActivity(t *testing.T) {
	columns := []string{"type", "at", "summary", "ref_id", "device_type"}

	tests := []struct {
		name     string
		queryErr error
		rows     [][]any
		want     []models.UserActivityEvent
		err      error
	}{
		{
			name: "returns the merged timeline",
			rows: [][]any{
				{"favorite_added", "2026-08-18T11:00:00Z", "added Algorithms to favorites", 9, ""},
				{"visit", "2026-08-18T10:00:00Z", "visited home from phone", 4, "phone"},
			},
			want: []models.UserActivityEvent{
				{Type: "favorite_added", At: "2026-08-18T11:00:00Z", Summary: "added Algorithms to favorites", RefID: 9},
				{Type: "visit", At: "2026-08-18T10:00:00Z", Summary: "visited home from phone", RefID: 4, DeviceType: "phone"},
			},
		},
		{
			name: "returns an empty timeline",
			rows: [][]any{},
			want: []models.UserActivityEvent{},
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			exp := mock.ExpectQuery(listUserTimelineQuery).WithArgs(7, 10, 0)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				rows := sqlmock.NewRows(columns)
				for _, row := range tt.rows {
					rows.AddRow(driverValues(row)...)
				}
				exp.WillReturnRows(rows)
			}

			got, err := repo.ListActivity(context.Background(), 7, 10, 0)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserRepository_GetLastDeviceType(t *testing.T) {
	tests := []struct {
		name     string
		queryErr error
		rows     *sqlmock.Rows
		want     string
		err      error
	}{
		{
			name: "returns the latest classified visit",
			rows: sqlmock.NewRows([]string{"device_type"}).AddRow("laptop"),
			want: "laptop",
		},
		{
			name:     "no classified visits",
			queryErr: sql.ErrNoRows,
			want:     "",
		},
		{
			name:     "query error",
			queryErr: errs.ErrDatabaseDown,
			err:      errs.ErrDatabaseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newTestUserRepo(t)

			exp := mock.ExpectQuery(getLastDeviceTypeQuery).WithArgs(7)
			if tt.queryErr != nil {
				exp.WillReturnError(tt.queryErr)
			} else {
				exp.WillReturnRows(tt.rows)
			}

			got, err := repo.GetLastDeviceType(context.Background(), 7)
			if tt.err != nil {
				assertRepoErr(t, mock, err, tt.err)
				return
			}
			assertRepoErr(t, mock, err, nil)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
