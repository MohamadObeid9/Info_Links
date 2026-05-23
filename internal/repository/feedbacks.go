package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

const (
	deleteFeedbackQuery = `DELETE FROM feedback WHERE id = $1`
	updateFeedbackQuery = `UPDATE feedback SET status = $1 WHERE id = $2`
	insertFeedbackQuery = `INSERT INTO feedback (category, rating, message) VALUES ($1, $2, $3)`

	listFeedbackBaseQuery        = `SELECT id, category, rating, message, status, created_at FROM feedback`
	listFeedbackNoFilterQuery    = listFeedbackBaseQuery + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	listFeedbackWithQQuery       = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithStatusQuery  = listFeedbackBaseQuery + ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	listFeedbackWithQStatusQuery = listFeedbackBaseQuery + ` WHERE (category ILIKE $1 OR message ILIKE $1) AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
)

type FeedbackRepository interface {
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, status string, id int) error
	Create(ctx context.Context, feedback models.Feedback) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error)
}

type PostgresFeedbackRepository struct {
	db *sql.DB
}

func NewPostgresFeedbackRepository(db *sql.DB) *PostgresFeedbackRepository {
	return &PostgresFeedbackRepository{db: db}
}

func (r *PostgresFeedbackRepository) Create(ctx context.Context, feedback models.Feedback) error {
	if _, err := r.db.ExecContext(ctx, insertFeedbackQuery, feedback.Category, feedback.Rating, feedback.Message); err != nil {
		return fmt.Errorf("insert feedback: %w", err)
	}
	return nil
}

func (r *PostgresFeedbackRepository) Update(ctx context.Context, status string, id int) error {
	res, err := r.db.ExecContext(ctx, updateFeedbackQuery, status, id)
	if err != nil {
		return fmt.Errorf("update feedback: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update feedback rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrFeedbackNotFound
	}
	return nil
}

func (r *PostgresFeedbackRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteFeedbackQuery, id)
	if err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete feedback rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrFeedbackNotFound
	}
	return nil
}

func (r *PostgresFeedbackRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error) {
	var (
		query string
		args  []any
	)
	switch {
	case q != "" && status != "":
		query = listFeedbackWithQStatusQuery
		args = []any{"%" + q + "%", status, limit, offset}
	case q != "":
		query = listFeedbackWithQQuery
		args = []any{"%" + q + "%", limit, offset}
	case status != "":
		query = listFeedbackWithStatusQuery
		args = []any{status, limit, offset}
	default:
		query = listFeedbackNoFilterQuery
		args = []any{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list feedbacks query: %w", err)
	}
	defer rows.Close()

	var feedbacks []models.Feedback
	for rows.Next() {
		var feedback models.Feedback
		if err := rows.Scan(&feedback.ID, &feedback.Category, &feedback.Rating, &feedback.Message, &feedback.Status, &feedback.CreatedAt); err != nil {
			return nil, fmt.Errorf("list feedbacks rows scan: %w", err)
		}
		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list feedbacks rows err: %w", err)
	}

	return feedbacks, nil
}
