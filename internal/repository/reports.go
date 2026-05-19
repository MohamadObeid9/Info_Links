package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type ReportRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, report models.Report) error
	Update(ctx context.Context, status string, id int) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error)
}

type PostgresReportRepository struct {
	db *sql.DB
}

const (
	insertReportQuery = `INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)`
	deleteReportQuery = `DELETE FROM reports WHERE id = $1`
	updateReportQuery = `UPDATE reports SET status = $1 WHERE id = $2`

	listReportsBaseQuery        = `SELECT id, course_name, link_url, description, status, created_at FROM reports`
	listReportsNoFilterQuery    = listReportsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listReportsWithQQuery       = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithStatusQuery  = listReportsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listReportsWithQStatusQuery = listReportsBaseQuery + ` WHERE (course_name ILIKE $1 OR description ILIKE $1 OR link_url ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

func NewPostgresReportRepository(db *sql.DB) *PostgresReportRepository {
	return &PostgresReportRepository{db: db}
}

func (r *PostgresReportRepository) Create(ctx context.Context, report models.Report) error {
	if _, err := r.db.ExecContext(ctx, insertReportQuery, report.CourseName, report.LinkURL, report.Description); err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *PostgresReportRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error) {
	var (
		query string
		args  []any
	)
	switch {
	case q != "" && status != "":
		query = listReportsWithQStatusQuery
		args = []any{"%" + q + "%", status, limit, offset}
	case q != "":
		query = listReportsWithQQuery
		args = []any{"%" + q + "%", limit, offset}
	case status != "":
		query = listReportsWithStatusQuery
		args = []any{status, limit, offset}
	default:
		query = listReportsNoFilterQuery
		args = []any{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports query: %w", err)
	}
	defer rows.Close()

	var reps []models.Report
	for rows.Next() {
		var rep models.Report
		if err := rows.Scan(&rep.ID, &rep.CourseName, &rep.LinkURL, &rep.Description, &rep.Status, &rep.CreatedAt); err != nil {
			return nil, fmt.Errorf("list reports rows scan: %w", err)
		}
		reps = append(reps, rep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list reports rows err: %w", err)
	}

	return reps, nil
}

func (r *PostgresReportRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteReportQuery, id)
	if err != nil {
		return fmt.Errorf("delete report: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete report rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrReportNotFound
	}
	return nil
}

func (r *PostgresReportRepository) Update(ctx context.Context, status string, id int) error {
	res, err := r.db.ExecContext(ctx, updateReportQuery, status, id)
	if err != nil {
		return fmt.Errorf("update report: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update report rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrReportNotFound
	}
	return nil
}
