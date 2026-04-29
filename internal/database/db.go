package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // The postgres driver
)

// DB is our global database connection pool.
var DB *sql.DB

// InitDB reads the connection string and connects to Postgres.
func InitDB(logger *slog.Logger, dbURL string) error {

	// Open the connection pool using the pgx driver
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Ping the database
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	// Basic pool tuning for production stability
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	logger.Info("successfully connected to database")

	DB = db
	return nil
}
