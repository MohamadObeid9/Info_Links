package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

type postgresCourseRepository struct {
	db *sql.DB
}

func NewPostgresCourseRepository(db *sql.DB) CourseRepository {
	return &postgresCourseRepository{db: db}
}

func (r *postgresCourseRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteCourseQuery, id)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete course rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrCourseNotFound
	}
	return nil
}

func (r *postgresCourseRepository) DeletePlacement(ctx context.Context, courseID, placementID int) error {
	resp, err := r.db.ExecContext(ctx, deleteCoursePlacementQuery, placementID, courseID)
	if err != nil {
		return fmt.Errorf("delete course placement: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete course placement rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrCourseNotFound
	}
	if _, err := r.db.ExecContext(ctx, deleteOrphanCourseQuery, courseID); err != nil {
		return fmt.Errorf("delete orphan course: %w", err)
	}
	return nil
}

func (r *postgresCourseRepository) Create(ctx context.Context, course models.Course) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create course: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var courseID int
	err = tx.QueryRowContext(ctx, findCourseIDByCodeQuery, course.Code).Scan(&courseID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find course by code: %w", err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		insErr := tx.QueryRowContext(ctx, insertCanonicalCourseQuery, course.Name, course.Code, course.IsOptional).Scan(&courseID)
		if insErr != nil {
			mapped := mapCourseConstraint(insErr)
			if !errors.Is(mapped, errs.ErrCourseCodeTaken) {
				return fmt.Errorf("insert course: %w", mapped)
			}
			if findErr := tx.QueryRowContext(ctx, findCourseIDByCodeQuery, course.Code).Scan(&courseID); findErr != nil {
				return fmt.Errorf("find course by code after unique: %w", findErr)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, insertCoursePlacementQuery, courseID, course.SemesterID, course.DisplayOrder); err != nil {
		return fmt.Errorf("insert course placement: %w", mapCourseConstraint(err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create course: %w", err)
	}
	return nil
}

func (r *postgresCourseRepository) GetByID(ctx context.Context, id int) (models.Course, error) {
	var c models.Course
	err := r.db.QueryRowContext(ctx, getCourseByIDQuery, id).Scan(
		&c.ID, &c.Name, &c.Code, &c.IsOptional,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Course{}, errs.ErrCourseNotFound
		}
		return models.Course{}, fmt.Errorf("get course: %w", err)
	}
	return c, nil
}

func (r *postgresCourseRepository) Update(ctx context.Context, course models.Course, id int) error {
	resp, err := r.db.ExecContext(ctx, updateCourseQuery, course.Name, course.Code, course.IsOptional, id)
	if err != nil {
		return fmt.Errorf("update course: %w", mapCourseConstraint(err))
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update course rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrCourseNotFound
	}
	if course.PlacementID > 0 && course.SemesterID > 0 {
		presp, err := r.db.ExecContext(ctx, updateCoursePlacementQuery, course.SemesterID, course.DisplayOrder, course.PlacementID, id)
		if err != nil {
			return fmt.Errorf("update course placement: %w", mapCourseConstraint(err))
		}
		paffected, err := presp.RowsAffected()
		if err != nil {
			return fmt.Errorf("update course placement rows affected: %w", err)
		}
		if paffected == 0 {
			return errs.ErrCourseNotFound
		}
	}
	return nil
}

func mapCourseConstraint(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolationCode {
		return err
	}
	switch pgErr.ConstraintName {
	case "courses_code_lower_uidx":
		return errs.ErrCourseCodeTaken
	case "course_placements_course_semester_key":
		return errs.ErrCourseAlreadyInSemester
	default:
		return errs.ErrCourseAlreadyInSemester
	}
}
