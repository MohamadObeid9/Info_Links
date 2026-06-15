package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/models"
)

type postgresPageViewRepository struct {
	db *sql.DB
}

func NewPostgresPageViewRepository(db *sql.DB) PageViewRepository {
	return &postgresPageViewRepository{db: db}
}

func (r *postgresPageViewRepository) List(ctx context.Context) ([]models.PageView, error) {
	rows, err := r.db.QueryContext(ctx, GetPageViewQuery)
	if err != nil {
		return nil, fmt.Errorf("get page views: %w", err)
	}

	defer func() { _ = rows.Close() }()
	var views []models.PageView
	for rows.Next() {
		var v models.PageView
		if err := rows.Scan(&v.ID, &v.Page, &v.VisitedAt); err != nil {
			return nil, fmt.Errorf("list page views row scan: %w", err)
		}
		views = append(views, v)
	}
	return views, nil
}

func (r *postgresPageViewRepository) Create(ctx context.Context, pv models.PageView) error {
	if _, err := r.db.ExecContext(ctx, insertPageViewQuery, pv.Page); err != nil {
		return fmt.Errorf("insert page view: %w", err)
	}
	return nil
}
