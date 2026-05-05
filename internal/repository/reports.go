package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

type ReportRepository interface {
	Create(ctx context.Context, report models.Report) error
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
