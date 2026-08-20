package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

type postgresLinkClickRepository struct {
	db *sql.DB
}

func NewPostgresLinkClickRepository(db *sql.DB) LinkClickRepository {
	return &postgresLinkClickRepository{db: db}
}

func (r *postgresLinkClickRepository) List(ctx context.Context) ([]models.LinkClick, error) {
	rows, err := r.db.QueryContext(ctx, GetLinkClickQuery)
	if err != nil {
		return nil, fmt.Errorf("get link clicks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var clicks []models.LinkClick
	for rows.Next() {
		var click models.LinkClick
		var linkID, extraLinkID sql.NullInt64
		if err := rows.Scan(&click.ID, &linkID, &extraLinkID, &click.ClickedAt); err != nil {
			return nil, fmt.Errorf("list link clicks row scan: %w", err)
		}
		if linkID.Valid {
			id := int(linkID.Int64)
			click.LinkID = &id
		}
		if extraLinkID.Valid {
			id := int(extraLinkID.Int64)
			click.ExtraLinkID = &id
		}
		clicks = append(clicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list link clicks rows: %w", err)
	}
	return clicks, nil
}

func (r *postgresLinkClickRepository) Create(ctx context.Context, lc models.LinkClick) error {
	if _, err := r.db.ExecContext(ctx, insertLinkClickQuery, lc.LinkID, lc.ExtraLinkID, lc.UserID); err != nil {
		return fmt.Errorf("insert link click: %w", err)
	}
	return nil
}
