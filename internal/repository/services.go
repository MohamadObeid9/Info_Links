package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/models"
)

type postgresServiceRepository struct {
	db *sql.DB
}

func NewPostgresServiceRepository(db *sql.DB) ServiceRepository {
	return &postgresServiceRepository{db: db}
}

func scanService(row interface {
	Scan(dest ...any) error
}, withClicks bool) (models.Service, error) {
	var s models.Service
	var links []byte
	dest := []any{
		&s.ID, &s.Title, &s.OwnerName, &s.Category, &s.Emoji,
		&s.Description, &s.LogoURL, &s.Phone, &s.URL, &links,
		&s.Status, &s.Trial, &s.StartedAt, &s.ExpiresAt, &s.DisplayOrder,
		&s.CreatedAt, &s.UpdatedAt,
	}
	if withClicks {
		dest = append(dest, &s.Clicks)
	}
	if err := row.Scan(dest...); err != nil {
		return models.Service{}, err
	}
	if len(links) == 0 {
		s.Links = []models.ServiceLink{}
	} else if err := json.Unmarshal(links, &s.Links); err != nil {
		return models.Service{}, fmt.Errorf("unmarshal service links: %w", err)
	}
	return s, nil
}

func (r *postgresServiceRepository) List(ctx context.Context, limit int, offset int, q string) ([]models.Service, error) {
	var args []any
	query := listServicesWithClicksQuery
	if q != "" {
		args = append(args, "%"+q+"%", "%"+q+"%")
		query += listServicesSearchWhere
	}
	args = append(args, limit, offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)
	query += fmt.Sprintf(` ORDER BY s.display_order ASC, s.created_at DESC LIMIT $%d OFFSET $%d`, limitIdx, offsetIdx)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var services []models.Service
	for rows.Next() {
		s, err := scanService(rows, true)
		if err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list services rows: %w", err)
	}
	return services, nil
}

func (r *postgresServiceRepository) Get(ctx context.Context, id int) (models.Service, error) {
	s, err := scanService(r.db.QueryRowContext(ctx, getServiceByIDQuery, id), false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Service{}, errs.ErrServiceNotFound
		}
		return models.Service{}, fmt.Errorf("get service: %w", err)
	}
	return s, nil
}

func (r *postgresServiceRepository) Create(ctx context.Context, s models.Service) (int, error) {
	links, err := json.Marshal(s.Links)
	if err != nil {
		return 0, fmt.Errorf("marshal service links: %w", err)
	}
	var id int
	if err := r.db.QueryRowContext(ctx, insertServiceQuery,
		s.Title, s.OwnerName, s.Category, s.Emoji, s.Description, s.LogoURL, s.Phone, s.URL, string(links), s.Status, s.Trial, s.StartedAt, s.ExpiresAt, s.DisplayOrder,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert service: %w", err)
	}
	return id, nil
}

func (r *postgresServiceRepository) Update(ctx context.Context, s models.Service, id int) error {
	links, err := json.Marshal(s.Links)
	if err != nil {
		return fmt.Errorf("marshal service links: %w", err)
	}
	resp, err := r.db.ExecContext(ctx, updateServiceQuery,
		s.Title, s.OwnerName, s.Category, s.Emoji, s.Description, s.LogoURL, s.Phone, s.URL, string(links), s.Status, s.Trial, s.StartedAt, s.ExpiresAt, s.DisplayOrder, id,
	)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("update service rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrServiceNotFound
	}
	return nil
}

func (r *postgresServiceRepository) Delete(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, deleteServiceQuery, id)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete service rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrServiceNotFound
	}
	return nil
}

func (r *postgresServiceRepository) Renew(ctx context.Context, id int) error {
	resp, err := r.db.ExecContext(ctx, renewServiceQuery, id)
	if err != nil {
		return fmt.Errorf("renew service: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("renew service rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrServiceNotFound
	}
	return nil
}

func (r *postgresServiceRepository) FreezeExpired(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, freezeExpiredServicesQuery); err != nil {
		return fmt.Errorf("freeze expired services: %w", err)
	}
	return nil
}

func (r *postgresServiceRepository) SetStatus(ctx context.Context, id int, status string) error {
	resp, err := r.db.ExecContext(ctx, setServiceStatusQuery, status, id)
	if err != nil {
		return fmt.Errorf("set service status: %w", err)
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("set service status rows affected: %w", err)
	}
	if affected == 0 {
		return errs.ErrServiceNotFound
	}
	return nil
}

func (r *postgresServiceRepository) InsertClick(ctx context.Context, click models.ServiceClick) error {
	if _, err := r.db.ExecContext(ctx, insertServiceClickQuery, click.ServiceID, click.UserID, click.PageContext, click.LinkTarget, click.URL, click.DeviceType); err != nil {
		return fmt.Errorf("insert service click: %w", err)
	}
	return nil
}

func (r *postgresServiceRepository) GetClickCount(ctx context.Context, id int) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, countServiceClicksQuery, id).Scan(&count); err != nil {
		return 0, fmt.Errorf("count service clicks: %w", err)
	}
	return count, nil
}

// NewService creates a service with default trial dates if not provided.
func NewService(title, ownerName, category, emoji, description, logoURL, phone, url string, links []models.ServiceLink, status string, trial bool, durationDays int) models.Service {
	now := time.Now()
	expires := now.AddDate(0, 0, durationDays)
	return models.Service{
		Title:        title,
		OwnerName:    ownerName,
		Category:     category,
		Emoji:        emoji,
		Description:  description,
		LogoURL:      logoURL,
		Phone:        phone,
		URL:          url,
		Links:        links,
		Status:       status,
		Trial:        trial,
		StartedAt:    now.Format(time.RFC3339),
		ExpiresAt:    expires.Format(time.RFC3339),
		DisplayOrder: 0,
	}
}
