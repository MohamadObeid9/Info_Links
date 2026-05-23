package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

const (
	insertLinkClickQuery = `INSERT INTO link_clicks (link_id) VALUES ($1)`
	GetLinkClickQuery    = `SELECT id, link_id, clicked_at FROM link_clicks WHERE link_id IS NOT NULL ORDER BY clicked_at DESC`
)

type LinkClickRepository interface {
	List(ctx context.Context) ([]models.LinkClick, error)
	Create(ctx context.Context, lc models.LinkClick) error
}

type PostgresLinkClickRepository struct {
	db *sql.DB
}

func NewPostgresLinkClickRepository(db *sql.DB) *PostgresLinkClickRepository {
	return &PostgresLinkClickRepository{db: db}
}

func (r *PostgresLinkClickRepository) Create(ctx context.Context, lc models.LinkClick) error {
	if _, err := r.db.ExecContext(ctx, insertLinkClickQuery, lc.LinkID); err != nil {
		return fmt.Errorf("insert link click: %w", err)
	}
	return nil
}
func (r *PostgresLinkClickRepository) List(ctx context.Context) ([]models.LinkClick, error) {
	rows, err := r.db.QueryContext(ctx, GetLinkClickQuery)
	if err != nil {
		return nil, fmt.Errorf("get link clicks: %w", err)
	}

	defer rows.Close()
	var clicks []models.LinkClick
	for rows.Next() {
		var click models.LinkClick
		if err := rows.Scan(&click.ID, &click.LinkID, &click.ClickedAt); err != nil {
			return nil, fmt.Errorf("list link clicks row scan: %w", err)
		}
		clicks = append(clicks, click)
	}
	return clicks, nil
}
