# Info Links

Centralized course resource hub for CNAM Lebanon CS students — **50+ courses**, **300+ active users**, served by a **production Go API** .

[![CI](https://github.com/MohamadObeid9/Info_Links/actions/workflows/ci.yml/badge.svg)](https://github.com/MohamadObeid9/Info_Links/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**[infolinks.app](https://infolinks.app)** · [User guide](docs/user-guide.md) · [ADRs](docs/adr/) · [Contributing](CONTRIBUTING.md)

## Quick start

```bash
git clone https://github.com/MohamadObeid9/Info_Links.git && cd Info_Links
cp .env.example .env   # fill DATABASE_URL, JWT_SECRET, Supabase keys
go run ./cmd/server    # → http://localhost:8080
```

## Overview

Info Links centralizes course materials (Google Drive, Classroom, Telegram, and more) by program, year, and semester. The backend is Go (`net/http`, layered architecture, ~88% test coverage); the frontend is vanilla JS with Vite. Production runs as Docker on Render with CI-gated deploys.

Using the app? See the [user guide](docs/user-guide.md).

---

## Tech stack

| Layer | Technology |
|-------|------------|
| **Frontend** | HTML5, CSS3, Vanilla JavaScript, Vite |
| **Backend** | Go 1.25 (`net/http`, no web framework) |
| **Database** | PostgreSQL via Supabase |
| **Auth** | Supabase Auth at login → app-issued JWT for admin routes |
| **Observability** | `slog`, Prometheus `/metrics`, `/healthz`, `/readyz` |
| **Deployment** | Render (Docker) + Supabase |
| **CI** | GitHub Actions — test, lint, govulncheck, frontend build, `docker build` |

---

## Architecture

Layered backend: HTTP handlers stay thin; business rules live in services; SQL lives in repositories.

```text
HTTP request
  → middleware (CORS, request ID, metrics, rate limit, recovery, security headers)
  → handler (internal/api)
  → service (internal/service)
  → repository (internal/repository)
  → PostgreSQL (Supabase)
```

The Go server also:

- Serves the built frontend from `frontend/dist` (or `frontend/` source in local dev when `dist/` is absent)
- Renders SSR SEO pages for `/course/{code}`, `/program/{slug}`, `/courses`, sitemap, and robots.txt
- Exposes a JSON API under `/api/*` — see `GET /api` for a live endpoint directory

**Deploy pipeline:**

```text
PR → GitHub Actions (test, lint, govulncheck, docker build) → merge main → Render (checksPass) → infolinks.app
```

Deep dives: [`docs/adr/`](docs/adr/) · [`docs/learnings/`](docs/learnings/)

---

## Project structure

```text
info_links/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/              # Router, handlers, JSON helpers
│   ├── service/          # Validation and business logic
│   ├── repository/       # Postgres queries behind interfaces
│   ├── middleware/       # Auth, rate limit, metrics, logging, recovery
│   ├── seo/              # Server-side SEO page rendering
│   ├── config/           # Environment configuration
│   ├── database/         # DB client (pgx)
│   ├── models/           # Shared domain types
│   └── errs/             # Sentinel errors
├── frontend/             # Vanilla JS SPA (Vite for dev/build)
├── docs/
│   ├── adr/              # Architecture Decision Records
│   ├── learnings/        # Per-package engineering notes
│   └── user-guide.md     # Student and admin walkthrough
├── Dockerfile            # Multi-stage: Node build → Go build → distroless
├── render.yaml           # Render Blueprint (Docker runtime, CI-gated deploy)
├── docker-compose.yml    # Local container + file watch
├── .air.toml             # Go hot-reload config
└── .env.example          # Required environment variables
```

---

## Getting started

### Prerequisites

- **Git**
- **Go 1.25+** ([go.mod](go.mod))
- **Node.js 20+** (frontend build and Vite dev server)
- **Supabase project** with a Postgres connection string
- **Docker & Docker Compose v2.22+** (optional, for containerized dev)
- **[Air](https://github.com/air-verse/air)** (optional, for Go hot reload)

### Environment

Copy [`.env.example`](.env.example) to `.env` and fill in:

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Supabase Postgres connection string |
| `JWT_SECRET` | Yes | Secret for signing admin JWTs |
| `SUPABASE_URL` | Yes | Supabase project URL |
| `SUPABASE_ANON_KEY` | Yes | Supabase anon/public key (admin login) |
| `PORT` | No | Default `8080` |
| `APP_ENV` | No | `development` or `production` |
| `SITE_BASE_URL` | No | Canonical URL for SEO/sitemap |
| `CORS_ALLOWED_ORIGINS` | No | Comma-separated origins for local Vite dev |
| `METRICS_BASIC_AUTH_*` | Prod only | Required when `APP_ENV=production` |

### Run locally

**Fastest path** — no frontend build needed (serves `frontend/` source):

```bash
go run ./cmd/server
```

**Production-like assets** — includes PWA manifest and service worker:

```bash
cd frontend && npm ci && npm run build && cd ..
go run ./cmd/server
```

Open [http://localhost:8080](http://localhost:8080).

---

## Development

Pick the workflow that matches what you're changing.

| Goal | Command |
|------|---------|
| Fastest dev loop | `air` + `npm run dev` (see below) |
| API work, single port | `go run ./cmd/server` |
| Full stack in Docker, auto-rebuild | `docker compose up --build --watch` |
| Verify container builds once | `docker compose up --build` |
| Match CI image build | `docker build -t infolinks:local .` |
| Restart existing container | `docker compose up` |

### Native (recommended for daily coding)

**Terminal 1** — Go API on port 8080:

```bash
go run ./cmd/server
# or, for Go hot reload:
air
```

**Terminal 2** — Vite dev server on port 5173 (proxies `/api` and SEO routes to 8080):

```bash
cd frontend && npm ci && npm run dev
# → http://localhost:5173
```

Use **5173** while editing HTML/CSS/JS. Fastest feedback loop.

### Docker (production-like)

Requires Docker Compose v2.22+ (for `watch`).

**Active development in a container** — rebuilds the image when project files change (`develop.watch` in `docker-compose.yml`):

```bash
docker compose up --build --watch
# → http://localhost:8080
```

Slower than Air + Vite, but matches the production Docker layout without installing Go/Node locally.

**One-shot verify:**

```bash
docker compose up --build
```

**Build image only (same as CI):**

```bash
docker build -t infolinks:local .
docker run --rm -p 8080:8080 --env-file .env infolinks:local
```

---

## Testing & CI

```bash
go test -race ./...
golangci-lint run ./...
cd frontend && npm ci && npm run build
docker build -t infolinks:local .
```

CI runs on every push/PR: Go build, `go test -race` with coverage, golangci-lint, govulncheck, frontend `npm run build`, and `docker build`. See [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

`main` is branch-protected — all CI jobs must pass before merge. Render deploys only after checks pass (`autoDeployTrigger: checksPass` in [`render.yaml`](render.yaml)).

---

## Deployment

Single Docker web service on [Render](https://render.com), configured in [`render.yaml`](render.yaml):

| Setting | Value |
|---------|-------|
| **Runtime** | Docker (`dockerfilePath: ./Dockerfile`) |
| **Start** | `ENTRYPOINT ["/server"]` |
| **Health check** | `GET /readyz` |
| **Auto-deploy** | On merge to `main`, only when CI checks pass |
| **Domains** | [infolinks.app](https://infolinks.app), www.infolinks.app |

The [Dockerfile](Dockerfile) is multi-stage: Node builds `frontend/dist`, Go compiles the server, final image runs on distroless as non-root. CI runs the same `docker build` before Render deploys — what is tested is what ships.

Set environment variables in the Render dashboard (see [`.env.example`](.env.example)). Secrets use `sync: false` in `render.yaml`. In production, `APP_ENV=production` and metrics basic auth are required for `/metrics`.

---

## Documentation

| Resource | Description |
|----------|-------------|
| [`docs/adr/`](docs/adr/) | Architecture Decision Records |
| [`docs/learnings/`](docs/learnings/) | Engineering notes per package |
| [`docs/load-test.md`](docs/load-test.md) | k6 load test results |
| [`docs/user-guide.md`](docs/user-guide.md) | Student and admin walkthrough |
| [`docs/roadmap.md`](docs/roadmap.md) | Milestones and planned work |
| [`frontend/README.md`](frontend/README.md) | Frontend layout |
| [DEV — origin story](https://dev.to/mohamadobeid9/i-built-a-free-course-resource-platform-for-my-university-heres-the-real-story-1645) | How Info Links started |
| [DEV — Go rebuild](https://dev.to/mohamadobeid9/from-supabase-only-to-production-go-month-1-of-rebuilding-info-links-3a4p) | Backend migration write-up |

---

## Contributing

Contributions welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: [SECURITY.md](SECURITY.md).

---

## Connect

- **Live site** — [infolinks.app](https://infolinks.app/)
- **GitHub** — [MohamadObeid9/Info_Links](https://github.com/MohamadObeid9/Info_Links)
- **Telegram** — [@Info_Links9](https://t.me/Info_Links9)
- **LinkedIn** — [MohamadObeid9](https://www.linkedin.com/in/mohamadobeid9/)

---

## License

MIT — see [LICENSE](LICENSE).

Built by students, for students.
