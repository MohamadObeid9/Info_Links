package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

type ContributionRepository interface {
	Create(ctx context.Context, contribution models.Contribution) error
}

type PostgresContributionRepository struct {
	db *sql.DB
}

func NewPostgresContributionRepository(db *sql.DB) *PostgresContributionRepository {
	return &PostgresContributionRepository{db: db}
}

func (r *PostgresContributionRepository) Create(ctx context.Context, contribution models.Contribution) error {
	const query = `INSERT INTO contributions (course_name, link_url, note) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, contribution.CourseName, contribution.LinkURL, contribution.Note); err != nil {
		return fmt.Errorf("insert Contribution: %w", err)
	}
	return nil
}
