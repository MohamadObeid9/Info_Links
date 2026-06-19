# Architecture Decision Records

Short documents explaining **why** Info Links is built the way it is. Useful for onboarding, interviews, and future-you.

| ADR | Title | Summary |
|-----|-------|---------|
| [001](001-postgres-as-primary-database.md) | Postgres as primary database | Relational model fits the course tree; JSON aggregation for `/api/content` |
| [002](002-supabase-for-auth-and-hosted-postgres.md) | Supabase for auth and hosted Postgres | Managed Postgres + Auth; Go backend owns API and issues app JWT |
| [003](003-net-http-without-web-framework.md) | `net/http` without a web framework | Standard library routing and middleware; minimal dependencies |
| [004](004-service-and-repository-layers.md) | Service and repository layers | Testable split between HTTP, business logic, and SQL |
| [005](005-render-for-deployment.md) | Render for deployment | Single web service, `/readyz`, free tier friendly |
| [006](006-server-side-seo-pages.md) | Server-side SEO pages | HTML for crawlers alongside the client-rendered SPA |

## Format

Each ADR follows:

- **Status** — Accepted, superseded, etc.
- **Context** — Problem and constraints
- **Decision** — What we chose
- **Consequences** — Trade-offs (positive and negative)

## When to add a new ADR

Add one when you make a significant architectural choice that would be hard to infer from code alone — new queue, cache, auth provider, deployment target, etc.
