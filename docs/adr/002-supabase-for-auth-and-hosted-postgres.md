# ADR 002: Supabase for Auth and Hosted Postgres

## Status

Accepted

## Context

Info Links needs:

1. A **managed Postgres** instance reachable from Render without self-hosting a database server
2. **Admin authentication** — password verification and role management without building user management from scratch
3. **Low operational cost** — student project, no budget for dedicated DevOps

Alternatives considered:

- **Self-hosted Postgres on a VPS** — full control, but backups, patches, and uptime are on us
- **Auth0 / Clerk** — strong auth UX, extra cost and integration surface for a single admin role
- **Custom users table + bcrypt in Go** — full ownership, but password reset, email verification, and security maintenance add scope
- **Supabase Auth + Postgres** — hosted DB and auth in one platform with a generous free tier

## Decision

Use **Supabase** for:

- **Hosted PostgreSQL** — connection via `DATABASE_URL`; Go uses standard `database/sql`, not Supabase client SDK for queries
- **Admin login verification** — `POST /api/auth/login` calls Supabase Auth REST API with email/password; backend checks `app_metadata.role == "admin"` on the returned token
- **App-issued JWT for API access** — after Supabase validates credentials, the Go backend signs a separate JWT (`JWT_SECRET`, 7-day expiry, `admin: true` claim) used by `RequireAdmin` middleware on admin routes

The Go backend remains the **API authority**. Supabase is identity + database hosting, not the public API layer.

## Consequences

### Positive

- No database server to operate; Supabase handles backups and availability
- Admin passwords managed through Supabase dashboard / Auth, not stored in our schema
- Standard Postgres connection string — portable if we migrate off Supabase later
- Free tier sufficient for 300+ daily users at current scale

### Negative

- Two auth tokens in play (Supabase session token at login, app JWT afterward) — must explain clearly in interviews
- Tight coupling to Supabase Auth REST API in `auth_handlers.go` — switching providers requires rewriting login
- `DATABASE_URL` and Supabase env vars are required at startup — local dev needs a Supabase project or compatible Postgres + auth setup
- Row Level Security (RLS) on Supabase is available but **not** the primary access control — the Go backend connects with a service role connection string and enforces auth in middleware

## References

- `internal/api/auth_handlers.go` — Supabase login + app JWT issuance
- `internal/middleware/auth_middleware.go` — app JWT validation
- `internal/config/config.go` — `DATABASE_URL`, `SUPABASE_URL`, `SUPABASE_ANON_KEY`
- `docs/learnings/api.md` — auth flow section
