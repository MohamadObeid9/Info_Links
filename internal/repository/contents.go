package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ContentRepository interface {
	Get(ctx context.Context) ([]byte, error)
}

const getContentQuery = `
	WITH content AS (
		SELECT
			(SELECT COALESCE(json_agg(p ORDER BY display_order ASC), '[]') FROM programs p) as programs,
			(SELECT COALESCE(json_agg(y ORDER BY display_order ASC), '[]') FROM years y) as years,
			(SELECT COALESCE(json_agg(s ORDER BY display_order ASC), '[]') FROM semesters s) as semesters,
			(SELECT COALESCE(json_agg(c ORDER BY display_order ASC), '[]') FROM courses c) as courses,
			(SELECT COALESCE(json_agg(l ORDER BY display_order ASC), '[]') FROM links l WHERE course_id IS NOT NULL) as links,
			(SELECT COALESCE(json_agg(ex ORDER BY display_order ASC), '[]') FROM extra_sections ex) as extra_sections,
			(SELECT COALESCE(json_agg(el ORDER BY display_order ASC), '[]') FROM extra_links el) as extra_links
	)
	SELECT json_build_object(
		'programs', programs,
		'years', years,
		'semesters', semesters,
		'courses', courses,
		'links', links,
		'extra_sections', extra_sections,
		'extra_links', extra_links
	) FROM content;
`

type PostgresContentRepository struct {
	db *sql.DB
}

func NewPostgresContentRepository(db *sql.DB) *PostgresContentRepository {
	return &PostgresContentRepository{db: db}
}

func (c *PostgresContentRepository) Get(ctx context.Context) ([]byte, error) {
	var result string
	if err := c.db.QueryRowContext(ctx, getContentQuery).Scan(&result); err != nil {
		return nil, fmt.Errorf("get content: %w", err)
	}
	return []byte(result), nil
}
