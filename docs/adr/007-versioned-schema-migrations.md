# ADR 007: Versioned Schema Migrations

## Context

Info Links stores all application data in PostgreSQL on Supabase (ADR 001, ADR 002). The Go repositories in `internal/repository` assume a fixed set of tables — programs through link_clicks — with foreign keys, check constraints, and display-order columns.

Until this ADR, the **canonical schema lived only in the Supabase dashboard**. That meant:

- New contributors could not recreate the database from the repo
- Schema changes were invisible in code review
- CI ran repository tests with **sqlmock** only — no way to prove complex SQL (e.g. the `/api/content` JSON aggregation CTE) against real Postgres
- Disaster recovery or a new environment required manual reconstruction from memory or the dashboard

We needed schema to be **version-controlled, reviewable, and eventually applicable the same way in every environment** (local, CI, production).

Alternatives considered:

- **Dashboard-only changes** — zero setup, but not reproducible or reviewable; common senior-engineer red flag in repo reviews
- **Single `schema.sql` dump only** — good snapshot for reading and diffs, but Supabase exports are not always executable in dependency order on a blank database
- **Supabase CLI migrations** (`supabase/migrations/`) — natural if the whole workflow is Supabase-native; adds CLI coupling for a Go backend that already uses plain `DATABASE_URL`
- **golang-migrate** — numbered `.up.sql` / `.down.sql` files; fits the Go stack; works with any Postgres (Supabase prod, Docker CI container)
- **goose, Flyway, Liquibase** — viable; golang-migrate chosen for minimal deps and wide Go-community usage

## Decision

1. **Store the schema in git** under `db/migrations/`.

2. **Initial baseline:** `db/migrations/schema.sql` — a reference snapshot exported from the live Supabase public schema (June 2026). It documents all 13 application tables. The file header marks it as context-only (Supabase export); it is the source of truth for **review and diff**, not for blind `psql -f` on empty Postgres.

3. **Executable migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) numbered files in the same directory — `000001_initial_schema.up.sql` / `.down.sql` — bootstrap a fresh Postgres (local Docker, CI) with correct sequence and FK order.

4. **Change policy:** Any production schema change must update the repo (new numbered migration + refresh `schema.sql`) in the same PR cycle as the application code that depends on it. No dashboard-only edits.

5. **Supabase role unchanged:** Supabase remains hosted Postgres + Auth (ADR 002). Migrations apply via standard `DATABASE_URL`; admin users stay in Supabase Auth, not in app tables.

Operational docs live in [`db/README.md`](../../db/README.md).

## Consequences

- Schema structure is visible in git history and PR diffs — reviewers see column and constraint changes alongside Go code
- Onboarding no longer requires Supabase dashboard access just to understand tables
- Enables the planned integration-test job (Postgres service container + `migrate up`)
- Interview narrative: *"Schema changes are versioned, reviewed, and applied the same way in every environment"*
- Extra discipline when changing schema: new migration file + refresh snapshot, not only click in dashboard
- `000001` must not be re-run on existing Supabase prod — it targets empty databases; prod gets `000002+` only
- Must keep `schema.sql` in sync with prod after changes for easy human review
