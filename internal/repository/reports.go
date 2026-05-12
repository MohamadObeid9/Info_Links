package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func NewPostgresReportRepository(db *sql.DB) *PostgresReportRepository {
	return &PostgresReportRepository{db: db}
}

func (r *PostgresReportRepository) Create(ctx context.Context, report models.Report) error {
	const query = `INSERT INTO reports (course_name, link_url, description) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, report.CourseName, report.LinkURL, report.Description); err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *PostgresReportRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error) {
	query := "SELECT id, course_name, link_url, description, status, created_at FROM reports"
	var args []any
	argIdx := 1
	var conditions []string
	if q != "" {
		conditions = append(conditions, fmt.Sprintf("(course_name ILIKE $%d OR description ILIKE $%d OR link_url ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

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
	const query = `DELETE FROM reports WHERE id = $1`
	resp, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete report: %w", err) // err 500
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete report rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrReportNotFound // err 404
	}
	return nil
}

func (r *PostgresReportRepository) Update(ctx context.Context, status string, id int) error {
	const query = "UPDATE reports SET status = $1 WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, status, id)
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
