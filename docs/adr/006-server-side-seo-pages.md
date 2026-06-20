# ADR 006: Server-Side SEO Pages

## Context

Info Links is primarily a **client-rendered SPA** (Vanilla JS + Vite build). Students load course data via `GET /api/content` and navigate in the browser. That works for users but is weak for:

- **Search engines** — crawlers may not execute JS reliably; course pages would be invisible in Google
- **Social sharing** — link previews need `<title>`, meta description, and Open Graph tags in initial HTML
- **Discoverability** — students search course codes (e.g. "INF101 CNAM") on Google; we want `/course/INF101` to rank

Alternatives considered:

- **SPA-only + prerender service** — e.g. prerender.io or build-time static generation; extra service or build complexity
- **Frontend SSR (Next.js, etc.)** — would require rewriting the frontend stack
- **Meta tags injected client-side** — insufficient for most crawlers
- **Dedicated SEO microservice** — overkill for current scale
- **Go HTML handlers in the same binary** — reuse existing Postgres data and deployment unit

## Decision

Add **`internal/seo`** — server-rendered HTML in the same Go process:

| Route | Output |
|-------|--------|
| `/course/{code}` | Course page with links, meta, JSON-LD |
| `/program/{slug}` | Program listing |
| `/courses` | Course index |
| `/sitemap.xml` | XML sitemap for crawlers |
| `/robots.txt` | Crawl rules + sitemap URL |

Implementation:

- `seo.Handler` fetches data via `SEOService` → `SEORepository` (joined SQL, DTOs in `seo_models.go`)
- `render.go` uses `html/template` with escaped output
- `meta.go` / `jsonld.go` build titles, descriptions, schema.org graphs
- `slug.go` normalizes program names to URL slugs

Static file handler in `router.go` **excludes** SEO paths (`isSEOPath`) so they never fall through to SPA `index.html`.

## Consequences

- Crawlers receive full HTML without executing JavaScript
- JSON-LD (`Course`, `ItemList`) improves rich search results
- Sitemap drives indexation of 50+ course pages
- Same deployment as API — no extra service to operate
- Demonstrates full-stack thinking in interviews (API + SEO + SPA coexistence)
