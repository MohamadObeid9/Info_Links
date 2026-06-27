# Database schema

The Info Links Postgres schema is **versioned in this repo** instead of living only in the Supabase dashboard. That makes schema changes reviewable in PRs, reproducible across environments, and ready for CI integration tests against real Postgres.

See [ADR 007: Versioned schema migrations](../docs/adr/007-versioned-schema-migrations.md) for the rationale.

## Layout

```text
db/
└── migrations/
    ├── schema.sql                      # Reference snapshot from Supabase (read-only diff)
    ├── 000001_initial_schema.up.sql    # Executable bootstrap (golang-migrate)
    └── 000001_initial_schema.down.sql  # Rollback for dev/CI
```

- **`schema.sql`** — human-readable export for review and diffs; not meant to be executed directly.
- **`000001_*.sql`** — ordered migrations applied by [golang-migrate](https://github.com/golang-migrate/migrate) on fresh databases (local Docker, CI).

## Prerequisites

Install the migrate CLI:

```bash
# macOS
brew install golang-migrate

# Linux (curl)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

## Apply migrations

```bash
export DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable"

migrate -path db/migrations -database "$DATABASE_URL" up
```

Check version:

```bash
migrate -path db/migrations -database "$DATABASE_URL" version
```

Rollback one step (dev/CI only):

```bash
migrate -path db/migrations -database "$DATABASE_URL" down 1
```

**Production (Supabase):** existing databases already have this schema. Do not run `000001` against prod unless bootstrapping a new empty database. Future changes use `000002_*.up.sql`, etc.

## Tables

Application-owned tables (13):

| Table | Purpose |
|-------|---------|
| `programs` | Degree programs (e.g. Licence, Master) |
| `years` | Academic years within a program |
| `semesters` | Semesters within a year |
| `courses` | Courses with codes (e.g. INF101) |
| `links` | Resource links attached to courses |
| `extra_sections` | Non-course link groupings |
| `extra_links` | Links inside extra sections |
| `reports` | User-submitted broken-link reports |
| `contributions` | User-submitted new link suggestions |
| `feedback` | User feedback and ratings |
| `page_views` | Anonymous page visit analytics |
| `link_clicks` | Link click analytics |

Auth and admin users are managed by **Supabase Auth** (ADR 002), not in this schema.

## Changing the schema (policy)

1. **Do not** change production schema only in the Supabase dashboard without a matching repo update.
2. Create a new numbered migration:

   ```bash
   migrate create -ext sql -dir db/migrations -seq add_some_column
   ```

3. Edit the generated `.up.sql` and `.down.sql` files.
4. Apply locally, run tests, open a PR.
5. Apply to Supabase prod, then refresh `schema.sql` so the snapshot stays in sync.

For destructive changes (drop column, rename), coordinate with a deploy that stops reading the old shape first.

**Never edit** a migration file that has already been applied to production — add a new one instead.

## Refreshing `schema.sql` from Supabase

After prod schema changes, re-export for easy diffing:

```bash
pg_dump "$DATABASE_URL" \
  --schema-only \
  --schema=public \
  --no-owner \
  --no-privileges \
  > db/migrations/schema.sql
```

Add the Supabase warning comment at the top if missing. Open a PR with the diff.

## Local development

Most contributors use a **hosted Supabase** instance via `DATABASE_URL` in `.env` (see root [README](../README.md)).

For a **local Postgres** (integration tests, offline work):

```bash
docker run --rm -d --name infolinks-pg \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=infolinks \
  -p 5432:5432 postgres:16-alpine

export DATABASE_URL="postgres://postgres:postgres@localhost:5432/infolinks?sslmode=disable"
migrate -path db/migrations -database "$DATABASE_URL" up
```

## Related docs

- [ADR 001: Postgres as primary database](../docs/adr/001-postgres-as-primary-database.md)
- [ADR 002: Supabase for auth and hosted Postgres](../docs/adr/002-supabase-for-auth-and-hosted-postgres.md)
- [ADR 007: Versioned schema migrations](../docs/adr/007-versioned-schema-migrations.md)
