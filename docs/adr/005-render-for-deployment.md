# ADR 005: Render for Deployment

## Context

Info Links is a single Go binary that:

- Serves the REST API
- Serves built frontend static files (`frontend/dist`)
- Renders SEO HTML pages

Deployment constraints for a student-maintained production app:

- **Free or low cost** at current traffic (~300+ monthly users, spikes during finals)
- **Simple CI/CD** — push to GitHub, auto-deploy only when CI passes
- **Health checks** — platform must detect unhealthy instances
- **One origin is enough** — Cloudflare sits in front of that origin for TLS, static cache, and `/api/content`
- **Reproducible builds** — frontend and backend must ship together
- Developer in Lebanon — prefer platforms with straightforward payment and EU-friendly options when scaling

Alternatives considered:

- **Native Render build (`go build` only)** — fast to set up, but skipped the Vite frontend build unless scripted separately; CI and production could diverge
- **Fly.io / Railway** — good DX, similar single-service model; Render chosen first for simplicity and documented free tier
- **Vercel / Netlify** — excellent for static frontend, awkward for long-lived Go API in one repo
- **AWS ECS / GCP Cloud Run** — production-grade, overkill for operational learning curve at MVP stage
- **VPS (Hetzner, etc.)** — cheapest at scale, requires manual TLS, deploy scripts, and monitoring setup

## Decision

Deploy as a **single Render Web Service** using **Docker**:

- **Runtime:** `docker` — Render builds from [`Dockerfile`](../../Dockerfile) (Node frontend build → Go compile → distroless runtime)
- **Dashboard config:** Render web service uses the repo `Dockerfile`; env vars and secrets live in the Render dashboard (not a checked-in Blueprint)
- **Auto-deploy:** `autoDeployTrigger: checksPass` — Render deploys only after GitHub Actions CI succeeds on `main`
- **Branch protection:** `main` requires CI status checks (`backend`, `frontend`, `docker-build`) before merge
- **Environment variables:** set in Render dashboard (`DATABASE_URL`, `JWT_SECRET`, Supabase keys, `SITE_BASE_URL`, metrics auth, `APP_ENV`)
- **Readiness probe:** `GET /readyz` — pings Postgres; Render stops routing if DB is unreachable
- **Liveness:** `GET /healthz` — process up, no DB check
- **Graceful shutdown:** on `SIGTERM` (Render deploys), `http.Server.Shutdown` drains in-flight requests before the process exits and the DB pool closes
- **Logs:** JSON `slog` in production for Render log stream
- **Domains:** `infolinks.app`, `www.infolinks.app`
- **CDN:** Cloudflare in front of Render — hashed Vite assets and `GET /api/content` (`Cache-Control: public, max-age=60, stale-while-revalidate=600`). A cron request every 10 minutes keeps `/api/content` warm at the edge.

Database stays on **Supabase** (separate service) — not in the same Render service as a Postgres container.

CI (`.github/workflows/ci.yml`) runs tests, lint, govulncheck, frontend build, and `docker build` on every push and PR. The same Dockerfile is used locally (`docker compose`), in CI, and on Render.

## Consequences

- One deploy unit matches one immutable Docker image — frontend and backend always ship together
- CI gates both merge (branch protection) and deploy (`checksPass`) — broken code does not reach production
- `/readyz` integrates with Render health checks out of the box
- Graceful shutdown avoids cutting mid-request traffic on every deploy
- Distroless final stage minimizes attack surface; no shell or package manager in production
- Deploy settings live in the Render dashboard; the Dockerfile in git is still the source of truth for the image
- Docker builds are slower than native Go builds on first deploy, but layer caching speeds up subsequent deploys
- Local dev can still use Air + Vite or `docker compose up --build --watch` without affecting production deploy path
- Cloudflare + the warm-up cron absorb student traffic (Grafana p95 after 21 Aug 2026). Origin also keeps a 60s in-memory copy of `/api/content` (ADR 010) so a CDN bypass does not stampede Postgres.
