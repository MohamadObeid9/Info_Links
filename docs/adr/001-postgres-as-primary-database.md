# ADR 001: Postgres as Primary Database

## Status

Accepted

## Context

Info Links stores a hierarchical navigation tree (programs → years → semesters → courses → links), user submissions (reports, feedback, contributions), and analytics (page views, link clicks). The app serves:

- One large JSON payload for the SPA (`GET /api/content`)
- Filtered admin lists with search and pagination
- Joined data for server-rendered SEO pages

We needed a database that handles structured relations, supports complex read queries, and scales with student traffic during exam periods without operational overhead beyond a student project.

Alternatives considered:

- **SQLite** — simple locally, but awkward for Render + concurrent writes from many students
- **MongoDB** — flexible schema, but the domain is inherently relational; joins would move into application code
- **Firebase** — fast to prototype, vendor lock-in, less aligned with Go backend hiring signal

## Decision

Use **PostgreSQL** as the sole database, accessed from Go via `database/sql` and the **pgx** driver.

Schema is normalized: separate tables for programs, years, semesters, courses, links, extra sections, and user-submission tables. The `/api/content` endpoint uses a single query with `json_agg` and `json_build_object` to assemble the full tree in Postgres rather than N+1 round trips.

## Consequences

### Positive

- Foreign keys and constraints keep the course tree consistent
- Powerful SQL for admin filters (`ILIKE`, status filters, pagination) and SEO joins
- JSON aggregation gives one fast read for the frontend bootstrap
- Postgres is the default choice for Go backend roles — good portfolio signal
- Hosted Postgres available cheaply via Supabase (see ADR 002)

### Negative

- Schema changes require migrations (manual or tooling) — no schemaless flexibility
- Complex SEO queries are harder to read than ORM-generated SQL
- Connection pool tuning matters under load (configured in `internal/database`)

## References

- `internal/repository/queries.go` — `getContentQuery` JSON aggregation
- `docs/learnings/database.md`
- `docs/learnings/repository.md`
