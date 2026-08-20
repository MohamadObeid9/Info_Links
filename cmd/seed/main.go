// Command seed loads an admin backup JSON into a local Postgres database.
//
// It never reads DATABASE_URL from .env, so a production Supabase URL cannot
// be targeted by accident. Pass -dsn and -allow-remote only when you mean it.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultLocalDSN = "postgres://postgres:postgres@localhost:5432/infolinks?sslmode=disable"

func main() {
	file := flag.String("file", "db/test-data.json", "path to an admin backup JSON")
	dsn := flag.String("dsn", defaultLocalDSN, "Postgres connection string")
	allowRemote := flag.Bool("allow-remote", false, "allow a non-localhost DSN (dangerous)")
	flag.Parse()

	if err := run(*file, *dsn, *allowRemote); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run(file, dsn string, allowRemote bool) error {
	if err := guardDSN(dsn, allowRemote); err != nil {
		return err
	}

	backup, err := loadBackup(file)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := apply(ctx, db, backup); err != nil {
		return err
	}

	fmt.Printf("seeded %s\n", file)
	fmt.Printf("  programs=%d years=%d semesters=%d courses=%d links=%d extra_sections=%d extra_links=%d\n",
		len(backup.Programs),
		len(backup.Years),
		len(backup.Semesters),
		len(backup.Courses),
		len(backup.Links),
		len(backup.ExtraSections),
		len(backup.ExtraLinks),
	)
	if backup.skippedClicks > 0 {
		fmt.Printf("  skipped %d link_clicks (content only; old click rows fail the current check constraint)\n", backup.skippedClicks)
	}
	return nil
}

func guardDSN(dsn string, allowRemote bool) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse dsn: %w", err)
	}
	host := u.Hostname()
	local := host == "localhost" || host == "127.0.0.1"
	if local {
		return nil
	}
	if !allowRemote {
		return fmt.Errorf("dsn host %q is not localhost; pass -allow-remote if you really want to seed %s", host, host)
	}
	fmt.Fprintf(os.Stderr, "warning: seeding remote host %s\n", host)
	return nil
}
