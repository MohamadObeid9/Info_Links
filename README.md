# 📚 Info Links

A modern, open-source platform designed to help students at Le CNAM Lebanon's Computer Science Department discover and organize course materials and learning resources in one centralized, user-friendly hub.

**Live site:** [infolinks.app](https://infolinks.app/)

---

## Table of contents

- [What is Info Links?](#-what-is-info-links)
- [Features](#-features)
- [Tech stack](#-tech-stack)
- [Architecture](#-architecture)
- [Project structure](#-project-structure)
- [Getting started](#-getting-started)
- [Development workflows](#-development-workflows)
- [Testing & CI](#-testing--ci)
- [Deployment](#-deployment)
- [How to use](#-how-to-use)
- [Contributing](#-contributing)
- [Documentation](#-documentation)
- [Connect](#-connect-with-us)
- [License & acknowledgments](#-license)

---

## 🌟 What is Info Links?

**Info Links** started as a simple idea to help students find course materials more easily. Today, it's evolved into a comprehensive platform covering **50+ courses** with hundreds of curated links — making it an essential resource for students pursuing their License and Master's degrees in Computer Science at Le CNAM Lebanon.

### The Growth Story

📈 **From Humble Beginnings to Impact**

- Started with just **4 courses** covering basic materials
- Grew to **50+ courses** with hundreds of curated resources
- Serving **300+ students** in under a year
- **Open-sourced** to empower the community and encourage contributions
- **Phase 8:** migrated from Supabase-direct + GitHub Pages to a **production Go backend** on Render

### The Impact

- **50+ Courses Covered** — From foundational to advanced subjects
- **Multi-disciplinary Resources** — Supports both License and partial Master's program courses
- **Growing Community** — Trusted by 300+ students across semesters
- **Regularly Updated** — New resources added consistently
- **Telegram Channel** — Parallel channel at [@Info_Links9](https://t.me/Info_Links9) for real-time updates

---

## ✨ Features

### For Students

- 🔍 **Smart Search** — Find courses by name or code instantly with keyboard shortcut (`/` or `Ctrl+K`)
- 📋 **Organized by Program** — Sorted by year, semester, and specialization
- 🏷️ **Easy Navigation** — Filter courses and identify optional vs. mandatory classes
- 🔗 **Multiple Resource Types** — Google Drive, Google Classroom, Telegram, and more — each with a color-coded badge
- ⭐ **My Courses (Favorites)** — Star courses to save them locally in your browser for quick access
- 🏷️ **Content Type Labels** — See what each link contains at a glance: TD, Cours, Videos, Sessions, Exams
- 🔗 **Multiple Content Types** — Links can have multiple content categories (e.g., TD + Cours + Videos)
- 🌓 **Light/Dark Mode** — Comfortable viewing in any lighting, with automatic system detection and persistence
- 📱 **Fully Responsive** — Works seamlessly on desktop, tablet, and mobile
- 💬 **Report & Contribute** — Report broken links or submit new resources with link type identification
- ⭐ **Feedback System** — Rate the platform (1–5 stars) by category and share suggestions
- 🔍 **Deep Linking** — Hash-based routing for direct view access (e.g., `#report-submit`)
- 🌐 **Multi-language Notes** — Important announcements available in English, French, and Arabic
- ⌨️ **Keyboard Shortcuts** — `/` or `Ctrl+K` to search, `Esc` to close modals
- 📱 **PWA Support** — Installable as a Progressive Web App with service worker
- 🔎 **SEO pages** — Server-rendered course/program pages for search engines

### For Admins

- ➕ **Full Course Management** — Add, edit, delete, and organize courses with program/year/semester placement
- 🏷️ **Advanced Labeling** — Mark courses as optional or mandatory
- 🔁 **Sibling Course Detection** — Courses shared across programs auto-sync names, codes, and links
- 🔗 **Multi-Content Link Management** — Assign multiple content types per link (TD, Cours, Videos, Sessions, Exams)
- 📊 **Analytics Dashboard** — Daily visitors, 7/30/90-day ranges, top clicked links, JSON export
- ✅ **Smart Contribution Review** — Approve user-submitted links with grouped course selector and sibling detection
- 🚨 **Report Management** — Handle user reports and improve content quality
- 💬 **Feedback Management** — View and manage user feedback with star ratings by category
- 🔐 **Secure Admin Panel** — JWT auth via the Go API (Supabase credentials at login)
- 📦 **Extra Resources** — Manage additional resource sections beyond regular courses

### Content Type Legend

| Badge | Meaning |
|-------|---------|
| ✏️ TD | Travaux Dirigés (exercises/tutorials) |
| 📄 Cours | Course materials/lectures |
| 🎬 Videos | Video recordings |
| 📝 Exams | Exam papers and solutions |
| 📦 Other | Other types of content |

### Link Type Legend

| Badge | Meaning |
|-------|---------|
| **TG** | Telegram |
| **GD** | Google Drive |
| **GC** | Google Classroom |
| **OT** | Other / External |

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | HTML5, CSS3, Vanilla JavaScript, Vite (build/dev) |
| **Backend** | Go 1.25 (`net/http`, no web framework) |
| **Database** | PostgreSQL via Supabase |
| **Auth** | Supabase Auth at login → app-issued JWT for admin routes |
| **Observability** | `slog`, Prometheus `/metrics`, `/healthz`, `/readyz` |
| **Deployment** | Render (Go web service) + Supabase |
| **Containers** | Multi-stage Docker, Docker Compose with file watch |
| **CI** | GitHub Actions — `go test -race`, golangci-lint, govulncheck, frontend build |

---

## 🏗️ Architecture

The backend follows a layered design: HTTP handlers stay thin; business rules live in services; SQL lives in repositories.

```text
HTTP request
  → middleware (CORS, request ID, metrics, rate limit, recovery, security headers)
  → handler (internal/api)
  → service (internal/service)
  → repository (internal/repository)
  → PostgreSQL (Supabase)
```

The Go server also:

- Serves the built frontend from `frontend/dist` (or `frontend/` source files in local dev when `dist/` is absent)
- Renders server-side SEO HTML for `/course/{code}`, `/program/{slug}`, `/courses`, sitemap, and robots.txt
- Exposes a JSON API under `/api/*` — see `GET /api` for a live endpoint directory

For deeper detail, see [`docs/adr/`](docs/adr/) and [`docs/learnings/`](docs/learnings/).

---

## 📁 Project structure

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
│   └── blog/             # Dev.to post drafts
├── Dockerfile            # Multi-stage: Node build → Go build → distroless
├── docker-compose.yml    # Local container + file watch
├── .air.toml             # Go hot-reload config
└── .env.example          # Required environment variables
```

---

## 🚀 Getting started

### Prerequisites

- **Git**
- **Go 1.25+** ([go.mod](go.mod))
- **Node.js 20+** (frontend build and Vite dev server)
- **Supabase project** with a Postgres connection string (or compatible Postgres)
- **Docker & Docker Compose** (optional, for containerized dev)
- **[Air](https://github.com/air-verse/air)** (optional, for Go hot reload)

### 1. Clone and configure

```bash
git clone https://github.com/MohamadObeid9/Info_Links.git
cd Info_Links
cp .env.example .env
```

Edit `.env` with your values:

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Supabase Postgres connection string |
| `JWT_SECRET` | Yes | Secret for signing admin JWTs |
| `SUPABASE_URL` | Yes | Supabase project URL |
| `SUPABASE_ANON_KEY` | Yes | Supabase anon/public key (admin login) |
| `PORT` | No | Default `8080` |
| `APP_ENV` | No | `development` or `production` |
| `SITE_BASE_URL` | No | Canonical URL for SEO/sitemap (default `http://localhost:8080`) |
| `CORS_ALLOWED_ORIGINS` | No | Comma-separated origins for local Vite dev |
| `METRICS_BASIC_AUTH_*` | Prod only | Required when `APP_ENV=production` |

### 2. Run locally (simplest)

**Option A — source frontend (no Node build needed)**

The server serves files from `frontend/` when `frontend/dist/` does not exist:

```bash
go run ./cmd/server
```

Open [http://localhost:8080](http://localhost:8080).

**Option B — production-like static assets**

Build the frontend first, then start the server:

```bash
cd frontend && npm ci && npm run build && cd ..
go run ./cmd/server
```

The server prefers `frontend/dist/` when present (includes PWA manifest and service worker from Vite).

---

## 💻 Development workflows

Pick the workflow that matches what you're changing.

### Full stack, single port (backend + static files)

```bash
go run ./cmd/server
# → http://localhost:8080
```

Good for API work and quick frontend checks without Vite.

### Frontend hot reload (Vite) + backend API

Terminal 1 — Go API on port 8080:

```bash
go run ./cmd/server
# or, for Go hot reload:
air
```

Terminal 2 — Vite dev server on port 5173 (proxies `/api`, SEO routes to 8080):

```bash
cd frontend
npm ci
npm run dev
# → http://localhost:5173
```

Use **5173** while editing HTML/CSS/JS. API and SEO requests are proxied to the backend.

### Go hot reload with Air

From the repo root (requires [Air](https://github.com/air-verse/air) installed):

```bash
air
```

Rebuilds the Go binary on `.go` file changes. Does not rebuild the frontend — pair with `npm run dev` or rebuild `frontend/dist` when needed.

### Docker Compose with live rebuild

Builds the full image (frontend + Go) and rebuilds the container when project files change:

```bash
docker compose up --build --watch
# → http://localhost:8080
```

Uses `.env` from the repo root. Slower feedback loop than Air + Vite, but matches production layout without installing Go/Node locally.

---

## 🧪 Testing & CI

```bash
# All backend tests (race detector)
go test -race ./...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Frontend production build
cd frontend && npm ci && npm run build
```

CI runs on every push/PR: Go build, `go test -race` with coverage, golangci-lint, govulncheck, and frontend `npm run build`. See [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

---

## 🚢 Deployment

Info Links runs as a **single Go web service** on [Render](https://render.com):

| Setting | Value |
|---------|-------|
| **Build command** | `go build -o server ./cmd/server` |
| **Start command** | `./server` |
| **Health check** | `GET /readyz` |

Set all required environment variables from `.env.example`. In production, `METRICS_BASIC_AUTH_USER` and `METRICS_BASIC_AUTH_PASSWORD` are required for the `/metrics` endpoint.

The [Dockerfile](Dockerfile) builds the frontend with Node, compiles Go, and runs on distroless — suitable for any container host.

---

## 📖 How to Use

### For Students

1. **Browse Courses** — Start on the home page to see all available courses
2. **Search** — Use the search bar (or press `/`) to find specific courses
3. **Filter** — Use program tabs and year/semester filters to narrow down
4. **Access Resources** — Click on resource links (Google Drive, Telegram, Classroom, etc.)
5. **Save Favorites** — Click the ★ star on any course to save it to "My Courses"
6. **Report Issues** — Use the "Report" section to report broken links
7. **Contribute** — Submit new links with URL, link type, and notes
8. **Toggle Theme** — Switch between light and dark mode
9. **Provide Feedback** — Rate the platform by category and share suggestions

### For Admins

1. **Log In** — Access the admin panel via the "Admin" button
2. **Manage Courses** — Add, edit, or remove courses and resources
3. **Add Links** — Assign multiple content types to each link
4. **Sibling Sync** — When editing shared courses, choose to update all occurrences
5. **Review Contributions** — Approve user-submitted links
6. **View Analytics** — Track visitor data and top clicked links
7. **Export Data** — Download analytics as JSON
8. **Manage Feedback** — Review user feedback and ratings

---

## 🤝 Contributing

We welcome contributions — course resources, bug fixes, and backend improvements.

### Contribute resources (no code)

Use **Report / Contribute** in the live app at [infolinks.app](https://infolinks.app/).

### Contribute code

1. **Fork** the repository and create a branch:
   ```bash
   git checkout -b fix/your-change
   ```
2. **Set up** locally (see [Getting started](#-getting-started))
3. **Make your changes** and run tests:
   ```bash
   go test -race ./...
   ```
4. **Open a pull request** with:
   - What changed and why
   - How you tested it
   - Screenshots for UI changes

**Guidelines:**

- Match existing code style and layering (`api` → `service` → `repository`)
- Add or update tests for backend logic changes
- For architectural choices, add or update an ADR in [`docs/adr/`](docs/adr/)
- Keep PRs focused — one concern per PR when possible
- Do not commit `.env` or secrets

---

## 📚 Documentation

| Resource | Description |
|----------|-------------|
| [`docs/adr/`](docs/adr/) | Why Postgres, Supabase, `net/http`, Render, SEO pages, etc. |
| [`docs/learnings/`](docs/learnings/) | Engineering notes per package |
| [`docs/load-test.md`](docs/load-test.md) | k6 load test results |
| [`frontend/README.md`](frontend/README.md) | Frontend layout |
| [DEV post — Part 1](https://dev.to/mohamadobeid9/i-built-a-free-course-resource-platform-for-my-university-heres-the-real-story-1645) | Origin story |
| [DEV post — Month 1 Go rebuild](https://dev.to/mohamadobeid9/from-supabase-only-to-production-go-month-1-of-rebuilding-info-links-3a4p) | Backend migration write-up |

---

## 📞 Connect With Us

- **Live site** — [infolinks.app](https://infolinks.app/)
- **GitHub** — [MohamadObeid9/Info_Links](https://github.com/MohamadObeid9/Info_Links)
- **Telegram** — [@Info_Links9](https://t.me/Info_Links9)

---

## 📜 License

This project is **open source** under the [MIT License](LICENSE).

---

## 🙏 Acknowledgments

- **Supabase** — Managed Postgres infrastructure for free
- **Render** — Hosted the whole project on their free version tier
- **All Contributors** — Thank you for making this project a success!
- **The Student Community** — For the feedback, support, and belief in this project

---

## 📊 Project Milestones

| Phase | Achievement |
|-------|-------------|
| **Phase 1** | Started with 4 courses covering basics |
| **Phase 2** | Expanded to 25+ courses |
| **Phase 3** | Reached 50+ courses with multiple resources per course |
| **Phase 4** | Serving 300+ students in under a year |
| **Phase 5** | Launched new website for better UX |
| **Phase 6** | Open-sourced project for community contributions |
| **Phase 7** | Favorites, content types, analytics, and PWA support |
| **Phase 8** | Go backend with layered architecture, observability, CI, and SEO |

---

## 💡 Future Roadmap

- [x] Advanced filtering and categorization
- [x] Personalized bookmarks (My Courses / Favorites)
- [x] Multi-language support (EN/FR/AR notes)
- [x] Community rating system for resources (Feedback)
- [x] Offline mode support (PWA / Service Worker)
- [x] Production Go backend with tests and observability
- [ ] Link health checker worker (companion Go service)
- [ ] Mobile app (iOS/Android)
- [ ] Push notifications for new resources
- [ ] Course schedule integration

---

**Made with ❤️ for Le CNAM Lebanon Computer Science Students**
