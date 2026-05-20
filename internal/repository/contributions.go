package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

const (
	insertContributionQuery = `INSERT INTO contributions (course_name, link_url, link_type, note) VALUES ($1, $2, $3, $4)`
	deleteContributionQuery = `DELETE FROM contributions WHERE id = $1`
	updateContributionQuery = `UPDATE contributions SET status = $1 WHERE id = $2`

	listContributionsBaseQuery        = `SELECT id, course_name, link_url, link_type, note, status, created_at FROM contributions`
	listContributionsNoFilterQuery    = listContributionsBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listContributionsWithQQuery       = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR link_type ILIKE $1 OR note ILIKE $1 ) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithStatusQuery  = listContributionsBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listContributionsWithQStatusQuery = listContributionsBaseQuery + ` WHERE (course_name ILIKE $1 OR link_url ILIKE $1 OR link_type ILIKE $1 OR note ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

type ContributionRepository interface {
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, status string, id int) error
	Create(ctx context.Context, contribution models.Contribution) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error)
}

type PostgresContributionRepository struct {
	db *sql.DB
}

func NewPostgresContributionRepository(db *sql.DB) *PostgresContributionRepository {
	return &PostgresContributionRepository{db: db}
}

func (c *PostgresContributionRepository) Create(ctx context.Context, contribution models.Contribution) error {
	if _, err := c.db.ExecContext(ctx, insertContributionQuery, contribution.CourseName, contribution.LinkURL, contribution.LinkType, contribution.Note); err != nil {
		return fmt.Errorf("insert contribution: %w", err)
	}
	return nil
}

func (c *PostgresContributionRepository) Delete(ctx context.Context, id int) error {
	resp, err := c.db.ExecContext(ctx, deleteContributionQuery, id)
	if err != nil {
		return fmt.Errorf("delete contribution: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete contribution rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrContributionNotFound
	}
	return nil
}

func (c *PostgresContributionRepository) Update(ctx context.Context, status string, id int) error {
	res, err := c.db.ExecContext(ctx, updateContributionQuery, status, id)
	if err != nil {
		return fmt.Errorf("update contribution: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update contribution rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrContributionNotFound
	}
	return nil
}

func (c *PostgresContributionRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error) {
	var (
		query string
		args  []any
	)
	switch {
	case q != "" && status != "":
		query = listContributionsWithQStatusQuery
		args = []any{"%" + q + "%", status, limit, offset}
	case q != "":
		query = listContributionsWithQQuery
		args = []any{"%" + q + "%", limit, offset}
	case status != "":
		query = listContributionsWithStatusQuery
		args = []any{status, limit, offset}
	default:
		query = listContributionsNoFilterQuery
		args = []any{limit, offset}
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list contributions query: %w", err)
	}
	defer rows.Close()

	var contributions []models.Contribution
	for rows.Next() {
		var contribution models.Contribution
		if err := rows.Scan(&contribution.ID, &contribution.CourseName, &contribution.LinkURL, &contribution.LinkType, &contribution.Note, &contribution.Status, &contribution.CreatedAt); err != nil {
			return nil, fmt.Errorf("list contributions rows scan: %w", err)
		}
		contributions = append(contributions, contribution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list contributions rows err: %w", err)
	}

	return contributions, nil
}
