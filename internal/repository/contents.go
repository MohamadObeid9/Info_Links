package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type postgresContentRepository struct {
	db *sql.DB
}

func NewPostgresContentRepository(db *sql.DB) ContentRepository {
	return &postgresContentRepository{db: db}
}

func (c *postgresContentRepository) Get(ctx context.Context) ([]byte, error) {
	var result string
	if err := c.db.QueryRowContext(ctx, getContentQuery).Scan(&result); err != nil {
		return nil, fmt.Errorf("get content: %w", err)
	}
	return []byte(result), nil
}
