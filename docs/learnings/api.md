## API Layer in This Project

### What the `api` package does

The `internal/api` package is the HTTP boundary of the backend. It owns:

- Route registration and middleware wrapping
- JSON request/response handling
- HTTP status codes and error messages
- Admin auth wiring on protected routes

Handlers decode HTTP, call a service, and map results/errors back to JSON.


---

### File roles


| File                 | Responsibility                                                                      |
| -------------------- | ----------------------------------------------------------------------------------- |
| `handler.go`         | `Handler` struct, small service interfaces, `NewHandler` validation, `LoggerWithID` |
| `router.go`          | Route groups, middleware chain, CORS, static files, security headers                |
| `*_handlers.go`      | Thin HTTP adapters per domain (links, courses, reports, …)                          |
| `helpers_http.go`    | Shared JSON helpers, body decoding, pagination                                      |
| `system_handlers.go` | `/healthz`, `/readyz`, `/api` root directory                                        |
| `auth_handlers.go`   | Admin login via Supabase, issues app JWT                                            |


Handlers are split by domain so each file stays small and focused.

---

### Request lifecycle (example: `POST /api/admin/links`)

```text
Client
  → CORS
  → RequestID + request logging
  → Metrics (Prometheus)
  → Rate limit (per IP)
  → Panic recovery
  → Security headers (CSP, X-Frame-Options, …)
  → RequireAdmin (JWT check)
  → handleAdminPostLink
  → LinkService.Create(ctx, link)
  → LinkRepository.Create(ctx, link)
  → Postgres
```

**Middleware order in `NewRouter`:** inner handler is built first (mux + routes), then layers wrap outward. The request passes through CORS last (outermost) before hitting the mux.

Steps inside `handleAdminPostLink`:

1. `decodeJSONBody` parses JSON into `models.Link` (1 MB cap, unknown fields rejected)
2. `h.linkService.Create(r.Context(), link)` runs validation + persistence
3. On success → `201` with `{"status":"ok"}`
4. On error → `mapPostLinkErr` maps domain errors to HTTP status

---

### Public vs admin vs SEO routes

Three registration functions in `router.go`:

**Public (`registerPublicRoutes`)** — no JWT required:

- `GET /api/content` — main navigation JSON
- `POST /api/reports`, `/api/feedback`, `/api/page_views`, … — user submissions
- `POST /api/auth/login` — admin login, returns JWT
- `GET /healthz`, `GET /readyz` — probes for Render/load balancers
- `GET /metrics` — provide the metrics for **Prometheus** , protected by a username/password

**Admin (`registerAdminRoutes`)** — every route wrapped with `middleware.RequireAdmin`:

- CRUD on links, courses, extra sections/links
- List/update/delete reports, feedback, contributions
- Analytics: page views, link clicks

**SEO (`registerSEORoutes`)** — separate `seo.Handler`, returns HTML not JSON:

- `/course/{code}`, `/program/{slug}`, `/courses`
- `/sitemap.xml`, `/robots.txt`

SEO is separate because crawlers need server-rendered pages with meta tags and JSON-LD, not the SPA's client-side routing.

**Static files** — `mux.Handle("/", …)` serves the frontend. SPA paths (`/`, `/admin`, …) fall back to `index.html`. SEO paths are excluded so they always hit the SEO handler.

---

### Why handlers depend on interfaces

`handler.go` defines small interfaces like:

```go
type linkService interface {
    Create(ctx context.Context, link models.Link) error
    Update(ctx context.Context, link models.Link, idStr string) error
    Delete(ctx context.Context, idStr string) error
}
```

The handler only knows the methods it calls — not Postgres, not concrete `*service.LinkService`.

**Benefits:**

- Unit tests pass mock services without a database
- Handlers stay thin; business logic lives in `internal/service`
- Swapping implementations (e.g. caching layer) does not touch HTTP code

**Tradeoff:** more boilerplate in `handler.go` and `NewHandler` validation. Worth it for testability at this scale.

---

### HTTP helpers (`helpers_http.go`)


| Helper                  | What it does                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `decodeJSONBody`        | Max 1 MB body, `DisallowUnknownFields`, returns `false` + 400 on bad JSON             |
| `writeJSON`             | Sets `Content-Type: application/json`, encodes payload                                |
| `writeJSONError`        | Consistent `{"error":"..."}` shape; 5xx responses include `request_id` when available |
| `parsePaginationParams` | Reads `limit`, `offset`, `q` from query string; caps `limit` at 100                   |


Handlers should use these instead of  `json.NewDecoder` calls so behavior stays consistent.

---

### Error mapping pattern

Services return typed errors from `internal/errs`. Handlers map them with `errors.Is`:

```go
func mapPostLinkErr(h *Handler, w http.ResponseWriter, r *http.Request, err error) {
    switch {
    case errors.Is(err, errs.ErrLinkURLAndLabelRequired):
        writeJSONError(w, r, http.StatusBadRequest, "Link url and link label are required")
    default:
        h.LoggerWithID(r).Error("create link failed", "error", err)
        writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
    }
}
```

**Rules:**

- Known domain error → specific 4xx + safe message
- Unknown error → log with request ID, return generic 500 (never leak `err.Error()` to clients)

Each domain has its own `mapXxxErr` helper in the same `*_handlers.go` file.

---

### Auth flow

**Login (`POST /api/auth/login`):**

1. Decode email/password from JSON
2. Verify credentials with Supabase Auth API (`verifyWithSupabase`)
3. Check `app_metadata.role == "admin"` in the Supabase JWT (`isAdmin`)
4. Issue a short-lived app JWT signed with `JWT_SECRET` (7-day expiry, `admin: true` claim)
5. Return `{"token":"..."}` to the frontend

**Protected routes:**

- Frontend sends `Authorization: Bearer <app-jwt>` on admin requests
- `RequireAdmin` validates signature, expiry, and `admin` claim before calling the handler

Supabase handles password verification; the app JWT keeps admin API auth self-contained and fast.

---

### Health and readiness


| Endpoint       | Purpose   | Behavior                                                |
| -------------- | --------- | ------------------------------------------------------- |
| `GET /healthz` | Liveness  | Always `200` — process is up                            |
| `GET /readyz`  | Readiness | Pings DB with 5s timeout; `503` if Postgres unreachable |


Render uses `/readyz` to decide whether to send traffic. `/healthz` answers "is the process alive?" without touching the database.

---

### Observability touchpoints

- **`LoggerWithID(r)`** — enriches logs with the request ID set by middleware (see `logging.md`)
- **`GET /metrics`** — Prometheus handler; basic-auth in production when configured, denied otherwise
- **`Panic recovery`** — middleware catches panics, logs stack, returns 500

Middleware details live in `internal/middleware/` — the API layer only consumes them in `NewRouter`.

---

### Common mistakes to avoid

- Putting SQL or validation logic in handlers (old pattern — removed during refactor)
- Returning raw internal error strings on 500 responses
- Forgetting to pass `r.Context()` into service calls (breaks timeouts/cancellation)
- Adding admin routes without `RequireAdmin` wrapper
- Debugging "missing request ID in logs" without checking middleware order

