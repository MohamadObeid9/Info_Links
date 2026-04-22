package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib" // The postgres driver
)

// DB is our global database connection pool.
var DB *sql.DB

// InitDB reads the connection string and connects to Postgres.
func InitDB() {
	// 1. Load the .env file.
	// We try multiple paths to find the .env file depending on where the code is run from.
	paths := []string{".env", "../.env", "../../.env", "../../../.env"}
	loaded := false
	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Note: No .env file found. Using system environment variables.")
	}

	// 2. Get the connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("ERROR: DATABASE_URL is not set. Please add it to your .env file.")
	}

	// 3. Open the connection pool using the pgx driver
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to open database config:", err)
	}

	// 4. Ping the database
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to the database. Check your password and URL! Error:", err)
	}

	// 5. Basic pool tuning for production stability
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("✅ Successfully connected to Supabase Postgres database!")
	
	DB = db
}
