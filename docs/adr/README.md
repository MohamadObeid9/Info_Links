# Architecture Decision Records

Short documents explaining **why** Info Links is built the way it is. Useful for onboarding, interviews, and future-you.

| ADR | Title | Summary |
|-----|-------|---------|
| [001](001-postgres-as-primary-database.md) | Postgres as primary database | Relational model fits the course tree; JSON aggregation for `/api/content` |
| [002](002-supabase-for-auth-and-hosted-postgres.md) | Supabase for auth and hosted Postgres | Managed Postgres + Auth; Go backend owns API and issues app JWT |
| [003](003-net-http-without-web-framework.md) | `net/http` without a web framework | Standard library routing and middleware; minimal dependencies |
| [004](004-service-and-repository-layers.md) | Service and repository layers | Testable split between HTTP, business logic, and SQL |
| [005](005-render-for-deployment.md) | Render for deployment | Docker web service, CI-gated deploy, Cloudflare in front, `/readyz` |
| [006](006-server-side-seo-pages.md) | Server-side SEO pages | HTML for crawlers alongside the client-rendered SPA |
| [007](007-versioned-schema-migrations.md) | Versioned schema migrations | `schema.sql` snapshot + golang-migrate files in `db/migrations/` |
| [008](008-student-identity-without-passwords.md) | Student identity without passwords | Name + number identity, guest claim, student JWT separate from admin auth |
| [009](009-canonical-courses-and-placements.md) | Canonical courses and placements | One course per code; programs are offerings; links are shared |
| [010](010-in-memory-content-cache.md) | In-memory `/api/content` cache | 60s TTL + `singleflight` at origin; k6 p95 1.53 ms |

## Format

Each ADR follows:

- **Status** — Accepted, superseded, etc.
- **Context** — Problem and constraints
- **Decision** — What we chose
- **Consequences** — Trade-offs (positive and negative)

## When to add a new ADR

Add one when you make a significant architectural choice that would be hard to infer from code alone — new queue, cache, auth provider, deployment target, etc.
