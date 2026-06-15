package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresExtraLinkRepository struct {
	db *sql.DB
}

func NewPostgresExtraLinkRepository(db *sql.DB) ExtraLinkRepository {
	return &postgresExtraLinkRepository{db: db}
}

func (r *postgresExtraLinkRepository) List(ctx context.Context) ([]models.ExtraLink, error) {
	rows, err := r.db.QueryContext(ctx, listExtraLinksQuery)
	if err != nil {
		return nil, fmt.Errorf("list extra links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []models.ExtraLink
	for rows.Next() {
		var l models.ExtraLink
		if err := rows.Scan(&l.ID, &l.SectionID, &l.Type, &l.URL, &l.Label, &l.Note, &l.ContentType, &l.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scan extra link: %w", err)
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list extra links rows: %w", err)
	}
	return links, nil
}

func (r *postgresExtraLinkRepository) Create(ctx context.Context, link models.ExtraLink) error {
	if _, err := r.db.ExecContext(ctx, insertExtraLinkQuery, link.SectionID, link.Type, link.URL, link.Label, link.Note, link.ContentType, link.DisplayOrder); err != nil {
		return fmt.Errorf("insert extra link: %w", err)
	}
	return nil
}

func (r *postgresExtraLinkRepository) Update(ctx context.Context, link models.ExtraLink, id int) error {
	resp, err := r.db.ExecContext(ctx, updateExtraLinkQuery, link.Type, link.URL, link.Label, link.Note, link.ContentType, id)
	if err != nil {
		return fmt.Errorf("update extra link: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update extra link rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrExtraLinkNotFound
	}
	return nil
}

func (r *postgresExtraLinkRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteExtraLinkQuery, id)
	if err != nil {
		return fmt.Errorf("delete extra link: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete extra link rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrExtraLinkNotFound
	}
	return nil
}
