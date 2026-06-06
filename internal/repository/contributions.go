package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresContributionRepository struct {
	db *sql.DB
}

func NewPostgresContributionRepository(db *sql.DB) ContributionRepository {
	return &postgresContributionRepository{db: db}
}

func (c *postgresContributionRepository) Delete(ctx context.Context, id int) error {
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

func (c *postgresContributionRepository) Update(ctx context.Context, status string, id int) error {
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

func (c *postgresContributionRepository) Create(ctx context.Context, contribution models.Contribution) error {
	if _, err := c.db.ExecContext(ctx, insertContributionQuery, contribution.CourseName, contribution.LinkURL, contribution.Note); err != nil {
		return fmt.Errorf("insert contribution: %w", err)
	}
	return nil
}

func (c *postgresContributionRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error) {
	query, args := buildFilteredListQuery(listQueries{
		noFilter:    listContributionsNoFilterQuery,
		withQ:       listContributionsWithQQuery,
		withStatus:  listContributionsWithStatusQuery,
		withQStatus: listContributionsWithQStatusQuery,
	}, limit, offset, q, status)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list contributions query: %w", err)
	}
	defer rows.Close()

	var contributions []models.Contribution
	for rows.Next() {
		var contribution models.Contribution
		if err := rows.Scan(&contribution.ID, &contribution.CourseName, &contribution.LinkURL, &contribution.Note, &contribution.Status, &contribution.CreatedAt); err != nil {
			return nil, fmt.Errorf("list contributions rows scan: %w", err)
		}
		contributions = append(contributions, contribution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list contributions rows err: %w", err)
	}

	return contributions, nil
}
