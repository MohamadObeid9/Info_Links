package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	DB *sql.DB
}

func New(dbUrl string, logger *slog.Logger) (*Client, error) {

	if logger == nil {
		return nil, fmt.Errorf("database logger is required")
	}

	if dbUrl == "" {
		return nil, fmt.Errorf("database_url is required")
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("successfully connected to database")

	return &Client{DB: db}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.DB == nil {
		return fmt.Errorf("database client is nil")
	}
	return c.DB.PingContext(ctx)
}

func (c *Client) Close() error {
	if c == nil || c.DB == nil {
		return nil
	}
	return c.DB.Close()
}
