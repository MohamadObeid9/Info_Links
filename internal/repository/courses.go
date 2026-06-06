package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
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

func (r *postgresCourseRepository) Create(ctx context.Context, course models.Course) error {
	if _, err := r.db.ExecContext(ctx, insertCourseQuery, course.SemesterID, course.Name, course.Code, course.IsOptional, course.DisplayOrder); err != nil {
		return fmt.Errorf("insert course: %w", err)
	}
	return nil
}

func (r *postgresCourseRepository) Update(ctx context.Context, course models.Course, id int) error {
	resp, err := r.db.ExecContext(ctx, updateCourseQuery, course.Name, course.Code, course.SemesterID, id)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update course rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrCourseNotFound
	}
	return nil
}
