# Baseline Snapshot

## Environment

- Date: 2026-04-21
- Backend command: `go test ./...`
- Frontend command: `npm run build` (not executable in current environment: `npm` missing)

## Backend Baseline

- `go test ./...` currently fails in `internal/api` because tests require live database connectivity.
- Failure mode: DNS/network resolution to Supabase host while running test suite.
- Current tests are integration-coupled via `TestMain -> database.InitDB()`.

## Frontend Baseline

- Build tool not runnable locally in this session due to missing Node/NPM binary.
- Static code audit confirms syntax/runtime risks in several JS modules and mixed module strategy.

## Endpoint Inventory

### Public

- `GET /api`
- `GET /api/`
- `GET /api/content`
- `GET /api/health`
- `POST /api/auth/login`
- `POST /api/page_views`
- `POST /api/link_clicks`
- `POST /api/reports`
- `POST /api/contributions`
- `POST /api/feedback`

### Admin

- `GET /api/admin/reports`
- `PATCH /api/admin/reports/{id}`
- `DELETE /api/admin/reports/{id}`
- `GET /api/admin/feedback`
- `PATCH /api/admin/feedback/{id}`
- `DELETE /api/admin/feedback/{id}`
- `GET /api/admin/contributions`
- `PATCH /api/admin/contributions/{id}`
- `DELETE /api/admin/contributions/{id}`
- `POST /api/admin/courses`
- `PATCH /api/admin/courses/{id}`
- `DELETE /api/admin/courses/{id}`
- `POST /api/admin/links`
- `PATCH /api/admin/links/{id}`
- `DELETE /api/admin/links/{id}`
- `GET /api/admin/page_views`
- `GET /api/admin/link_clicks`

## Immediate Priorities

1. Restore frontend runtime integrity and module consistency.
2. Fix auth correctness and API validation safeguards.
3. Align build output, static serving, and PWA assets.
4. Decouple tests from hard DB dependency and add CI quality gates.
