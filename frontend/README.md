# InfoLinks Frontend

Vanilla HTML, CSS, and JavaScript UI for Info Links. The Go server serves these files (or `dist/` after a Vite build).

## Structure

- `index.html` — main SPA shell
- `main.js` — bootstrap and event wiring
- `js/` — feature modules (`data.js`, `home.js`, `admin.js`, `services.js`, …)
- `styles/` — CSS (`app.css`, `admin.css`, `services.css`, `responsive.css`, …)
- `public/` — static assets copied as-is into the build (favicon, PWA icons under `public/assets/`)

## Features

- Course browsing, search, favorites, report/contribute/feedback flows
- Community services cards (sidebar, list intersperse, dedicated community view)
- Admin panel (courses, services, analytics, students, inbox)
- Light/dark theme and mobile layouts
- PWA manifest + service worker in production builds

## Development

**With Vite** (from `frontend/`, proxies `/api` and SEO routes to the Go server on `:8080`):

```bash
npm ci && npm run dev
# → http://localhost:5173
```

**Served by Go** (no frontend build; uses source files when `dist/` is absent):

```bash
# from repo root
go run ./cmd/server
# → http://localhost:8080
```

**Checks:**

```bash
npm run lint
npm test
npm run build
```

API calls use relative paths (`/api/...`).
