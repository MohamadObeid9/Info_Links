package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

const (
	deleteLinkQuery = `DELETE FROM links WHERE id = $1`
	updateLinkQuery = `UPDATE links SET type = $1, url = $2, label = $3, note = $4, content_type = $5 WHERE id = $6`
	insertLinkQuery = `INSERT INTO links (course_id, type, url, label, note, content_type, display_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

type LinkRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, link models.Link) error
	Update(ctx context.Context, link models.Link, id int) error
}

type PostgresLinkRepository struct {
	db *sql.DB
}

func NewPostgresLinkRepository(db *sql.DB) *PostgresLinkRepository {
	return &PostgresLinkRepository{db: db}
}

func (r *PostgresLinkRepository) Create(ctx context.Context, link models.Link) error {
	if _, err := r.db.ExecContext(ctx, insertLinkQuery, link.CourseID, link.Type, link.URL, link.Label, link.Note, link.ContentType, link.DisplayOrder); err != nil {
		return fmt.Errorf("insert link: %w", err)
	}
	return nil
}

func (r *PostgresLinkRepository) Delete(ctx context.Context, id int) error {
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

func (r *PostgresLinkRepository) Update(ctx context.Context, link models.Link, id int) error {
	resp, err := r.db.ExecContext(ctx, updateLinkQuery, link.Type, link.URL, link.Label, link.Note, link.ContentType, id)
	if err != nil {
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
