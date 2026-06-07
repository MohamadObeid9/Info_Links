package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresExtraSectionRepository struct {
	db *sql.DB
}

func NewPostgresExtraSectionRepository(db *sql.DB) ExtraSectionRepository {
	return &postgresExtraSectionRepository{db: db}
}

func (r *postgresExtraSectionRepository) List(ctx context.Context) ([]models.ExtraSection, error) {
	rows, err := r.db.QueryContext(ctx, listExtraSectionsQuery)
	if err != nil {
		return nil, fmt.Errorf("list extra sections: %w", err)
	}
	defer rows.Close()

	var sections []models.ExtraSection
	for rows.Next() {
		var s models.ExtraSection
		if err := rows.Scan(&s.ID, &s.Title, &s.Icon, &s.DisplayOrder); err != nil {
			return nil, fmt.Errorf("scan extra section: %w", err)
		}
		sections = append(sections, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list extra sections rows: %w", err)
	}
	return sections, nil
}

func (r *postgresExtraSectionRepository) Create(ctx context.Context, section models.ExtraSection) error {
	if _, err := r.db.ExecContext(ctx, insertExtraSectionQuery, section.Title, section.Icon, section.DisplayOrder); err != nil {
		return fmt.Errorf("insert extra section: %w", err)
	}
	return nil
}

func (r *postgresExtraSectionRepository) Update(ctx context.Context, section models.ExtraSection, id int) error {
	resp, err := r.db.ExecContext(ctx, updateExtraSectionQuery, section.Title, section.Icon, id)
	if err != nil {
		return fmt.Errorf("update extra section: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update extra section rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrExtraSectionNotFound
	}
	return nil
}

func (r *postgresExtraSectionRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete extra section begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, deleteExtraLinksBySectionQuery, id); err != nil {
		return fmt.Errorf("delete extra section links: %w", err)
	}

	resp, err := tx.ExecContext(ctx, deleteExtraSectionQuery, id)
	if err != nil {
		return fmt.Errorf("delete extra section: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete extra section rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrExtraSectionNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete extra section commit: %w", err)
	}
	return nil
}
