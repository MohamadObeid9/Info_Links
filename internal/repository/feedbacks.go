package repository

import (
	"context"
	"database/sql"
	"fmt"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresFeedbackRepository struct {
	db *sql.DB
}

func NewPostgresFeedbackRepository(db *sql.DB) FeedbackRepository {
	return &postgresFeedbackRepository{db: db}
}

func (r *postgresFeedbackRepository) Delete(ctx context.Context, id int) error {
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

func (r *postgresFeedbackRepository) Update(ctx context.Context, status string, id int) error {
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

func (r *postgresFeedbackRepository) Create(ctx context.Context, feedback models.Feedback) error {
	if _, err := r.db.ExecContext(ctx, insertFeedbackQuery, feedback.Category, feedback.Rating, feedback.Message); err != nil {
		return fmt.Errorf("insert feedback: %w", err)
	}
	return nil
}

func (r *postgresFeedbackRepository) List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error) {
	query, args := buildFilteredListQuery(listQueries{
		noFilter:    listFeedbackNoFilterQuery,
		withQ:       listFeedbackWithQQuery,
		withStatus:  listFeedbackWithStatusQuery,
		withQStatus: listFeedbackWithQStatusQuery,
	}, limit, offset, q, status)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list feedbacks query: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
