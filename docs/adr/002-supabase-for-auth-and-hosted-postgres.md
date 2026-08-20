# ADR 002: Supabase for Auth and Hosted Postgres

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
- **App-issued JWT for API access** — after Supabase validates credentials, the Go backend signs a separate JWT (`JWT_SECRET`, 365-day expiry, `admin: true` claim) used by `RequireAdmin` middleware on admin routes

The Go backend remains the **API authority**. Supabase is identity + database hosting, not the public API layer.

## Consequences

- No database server to operate; Supabase handles backups and availability
- Admin passwords managed through Supabase dashboard / Auth, not stored in our schema
- Standard Postgres connection string — portable if we migrate off Supabase later
- Free tier sufficient for 300+ monthly users at current scale
