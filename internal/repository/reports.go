package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresReportRepository struct {
	db *sql.DB
}

func NewPostgresReportRepository(db *sql.DB) ReportRepository {
	return &postgresReportRepository{db: db}
}

func (r *postgresReportRepository) Delete(ctx context.Context, id int) error {
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

func (r *postgresReportRepository) Create(ctx context.Context, report models.Report) error {
	if _, err := r.db.ExecContext(ctx, insertReportQuery, report.CourseName, report.LinkURL, report.Description, report.UserID); err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *postgresReportRepository) Update(ctx context.Context, status string, id int) error {
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

func (r *postgresReportRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error) {
	query, args := buildFilteredListQuery(listQueries{
		noFilter:    listReportsNoFilterQuery,
		withQ:       listReportsWithQQuery,
		withStatus:  listReportsWithStatusQuery,
		withQStatus: listReportsWithQStatusQuery,
	}, limit, offset, q, status)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reps []models.Report
	for rows.Next() {
		var rep models.Report
		var userID sql.NullInt64
		if err := rows.Scan(&rep.ID, &rep.CourseName, &rep.LinkURL, &rep.Description, &rep.Status, &rep.CreatedAt, &userID); err != nil {
			return nil, fmt.Errorf("list reports rows scan: %w", err)
		}
		if userID.Valid {
			rep.UserID = int(userID.Int64)
		}
		reps = append(reps, rep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list reports rows err: %w", err)
	}

	return reps, nil
}
