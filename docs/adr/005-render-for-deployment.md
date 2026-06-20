# ADR 005: Render for Deployment

## Context

Info Links is a single Go binary that:

- Serves the REST API
- Serves built frontend static files (`frontend/dist`)
- Renders SEO HTML pages

Deployment constraints for a student-maintained production app:

- **Free or low cost** at current traffic (~300+ monthly users, spikes during finals)
- **Simple CI/CD** — push to GitHub, auto-deploy
- **Health checks** — platform must detect unhealthy instances
- **No separate frontend CDN required** initially — one service is easier to operate
- Developer in Lebanon — prefer platforms with straightforward payment and EU-friendly options when scaling

Alternatives considered:

- **Fly.io / Railway** — good DX, similar single-service model; Render chosen first for simplicity and documented free tier
- **Vercel / Netlify** — excellent for static frontend, awkward for long-lived Go API in one repo
- **AWS ECS / GCP Cloud Run** — production-grade, overkill for operational learning curve at MVP stage
- **VPS (Hetzner, etc.)** — cheapest at scale, requires manual TLS, deploy scripts, and monitoring setup

## Decision

Deploy as a **single Render Web Service**:

- Build: `go build -o server cmd/server/main.go` (or multi-stage Docker build for container deploy)
- Start: `./server`
- Environment variables set in Render dashboard (`DATABASE_URL`, `JWT_SECRET`, Supabase keys, `SITE_BASE_URL`, metrics auth in production)
- **Readiness probe:** `GET /readyz` — pings Postgres; Render stops routing if DB is unreachable
- **Liveness:** `GET /healthz` — process up, no DB check
- Logs: JSON `slog` in production for Render log stream

Database stays on **Supabase** (separate service) — not in the same Render service as Postgres container.

## Consequences

- One deploy unit matches one binary — mental model stays simple
- GitHub integration for automatic deploys on push
- `/readyz` integrates with Render health checks out of the box
- Distroless Docker image option for minimal attack surface (`Dockerfile`)
