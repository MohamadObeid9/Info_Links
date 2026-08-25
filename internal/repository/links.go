package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresLinkRepository struct {
	db *sql.DB
}

func NewPostgresLinkRepository(db *sql.DB) LinkRepository {
	return &postgresLinkRepository{db: db}
}

func (r *postgresLinkRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteLinkQuery, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete link rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrLinkNotFound
	}
	return nil
}

func (r *postgresLinkRepository) Create(ctx context.Context, link models.Link) error {
	if _, err := r.db.ExecContext(ctx, insertLinkQuery, link.CourseID, link.Type, link.URL, link.Label, link.Note, link.ContentType, link.DisplayOrder); err != nil {
		if isUniqueViolation(err) {
			return errs.ErrLinkURLTaken
		}
		return fmt.Errorf("insert link: %w", err)
	}
	return nil
}

func (r *postgresLinkRepository) Update(ctx context.Context, link models.Link, id int) error {
	resp, err := r.db.ExecContext(ctx, updateLinkQuery, link.Type, link.URL, link.Label, link.Note, link.ContentType, id)
	if err != nil {
		if isUniqueViolation(err) {
			return errs.ErrLinkURLTaken
		}
		return fmt.Errorf("update link: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update link rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrLinkNotFound
	}
	return nil
}
