package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

const (
	deleteCourseQuery = `DELETE FROM courses WHERE id = $1`
	updateCourseQuery = `UPDATE courses SET name = $1, code = $2, semester_id = $3 WHERE id = $4`
	insertCourseQuery = `INSERT INTO courses (semester_id, name, code, is_optional, display_order) VALUES ($1, $2, $3, $4, $5)`
)

type CourseRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, course models.Course) error
	Update(ctx context.Context, course models.Course, id int) error
}

type PostgresCourseRepository struct {
	db *sql.DB
}

func NewPostgresCourseRepository(db *sql.DB) *PostgresCourseRepository {
	return &PostgresCourseRepository{db: db}
}

func (r *PostgresCourseRepository) Create(ctx context.Context, course models.Course) error {
	if _, err := r.db.ExecContext(ctx, insertCourseQuery, course.SemesterID, course.Name, course.Code, course.IsOptional, course.DisplayOrder); err != nil {
		return fmt.Errorf("insert course: %w", err)
	}
	return nil
}

func (r *PostgresCourseRepository) Delete(ctx context.Context, id int) error {
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

func (r *PostgresCourseRepository) Update(ctx context.Context, course models.Course, id int) error {
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
