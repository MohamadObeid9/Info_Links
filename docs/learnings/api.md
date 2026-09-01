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

- `GET /api/content` — main navigation JSON (`Cache-Control: public, max-age=60, stale-while-revalidate=600`; Cloudflare caches it in production)
- `GET /.well-known/api-catalog` — RFC 9727 API catalog (`application/linkset+json`)
- `GET /.well-known/oauth-protected-resource` — RFC 9728 PRM for student API auth
- `GET /.well-known/oauth-authorization-server` — AS metadata + `agent_auth` (anonymous guest → claim)
- `GET /.well-known/openid-configuration` — OIDC Discovery (same core fields, includes `jwks_uri`)
- `GET /.well-known/jwks.json` — JWKS (empty `keys`; tokens are HS256 shared-secret)
- `GET /.well-known/agent-card.json` — A2A Agent Card (HTTP+JSON skills for the public API)
- `GET /.well-known/agents-index.json` — DNS-AID / ANS-style org index (A2A + MCP cards for `info-links`)
- `GET /.well-known/agent-skills/index.json` — Agent Skills Discovery index (v0.2.0)
- `GET /.well-known/agent-skills/{name}/SKILL.md` — individual skill artifacts (+ digests in the index)
- `GET /.well-known/mcp/server-card.json` — MCP Server Card (SEP-1649 discovery)
- `GET /.well-known/http-message-signatures-directory` — Web Bot Auth JWKS (Ed25519), response signed with HTTP Message Signatures
- `GET|POST|DELETE /mcp` — advertised MCP Streamable HTTP endpoint (stub until full MCP is implemented)
- Browser **WebMCP** — `navigator.modelContext.registerTool` on homepage load (`frontend/js/webmcp.js`: search, programs, course lookup, navigate)
- `GET /auth.md` — Auth.md skill document for agents
- `GET /openapi.json` — OpenAPI 3.1 description (`service-desc`)
- `GET /api/docs` — human API docs in markdown (`service-doc`)
- `POST /api/reports`, `/api/feedback`, `/api/page_views`, … — user submissions
- `GET /api/services` — public community service listings (expired trials auto-freeze on read)
- `POST /api/service_clicks` — track a service card open (requires student JWT)
- `POST /api/auth/login` — admin login, returns JWT
- `GET /healthz`, `GET /readyz` — probes for Render/load balancers
- `GET /metrics` — provide the metrics for **Prometheus** , protected by a username/password

**Admin (`registerAdminRoutes`)** — every route wrapped with `middleware.RequireAdmin`:

- CRUD on links, courses, extra sections/links
- Community services CRUD plus renew / freeze / unfreeze
- List/update/delete reports, feedback, contributions
- Analytics: page views, link clicks, summary dashboards

**SEO (`registerSEORoutes`)** — separate `seo.Handler`, returns HTML not JSON:

- `/course/{code}`, `/program/{slug}`, `/courses`
- `/sitemap.xml`, `/robots.txt`

SEO is separate because crawlers need server-rendered pages with meta tags and JSON-LD, not the SPA's client-side routing.

**Static files** — `mux.Handle("/", …)` serves the frontend. SPA paths (`/`, `/admin`, …) fall back to `index.html`. SEO paths are excluded so they always hit the SEO handler. The homepage (`/`) also sends RFC 8288 `Link` headers (`api-catalog`, `service-desc`, `service-doc`, `describedby`) so agents can discover the API catalog without scraping HTML.

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
4. Issue an app JWT signed with `JWT_SECRET` (365-day expiry, `admin: true` claim)
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

### Caching (`GET /api/content` and static files)

Origin keeps a **process-local copy** of `GET /api/content` (60s TTL, `singleflight` on miss) so a flood that bypasses Cloudflare does not run the CTE once per request. Admin `GET /api/admin/content` always hits Postgres (`GetUncached`). Successful course/link/extra/service mutations call `Invalidate()`.

Cloudflare still sits in front of Render:

- **Hashed Vite assets** — `Cache-Control: public, max-age=31536000, immutable`
- **`GET /api/content`** — `public, max-age=3600, stale-while-revalidate=600`
- **Warm-up** — a cron GET every 10 minutes so the edge object does not expire while students are asleep

Grafana (Prometheus scrape of `/metrics`) showed p95/p99 for `/api/content` collapsing on 21 Aug 2026 after the CDN headers went live. A **cold** origin miss still runs the Postgres JSON aggregation once (`singleflight` shares that miss).

Origin-only k6 (2026-09-01, local Go + remote Supabase), after the in-memory cache:

| | Normal (50 VUs) | Burst (30 VUs, 10s) |
|---|---|---|
| Throughput | 39.9 req/s | **17,009 req/s** |
| HTTP 200 | 100% | 0.07% (token bucket) |
| HTTP 429 | 0 | 99.93% |
| p95 | **1.53 ms** | 8.91 ms (200s only) |

Same day without the origin cache: normal p95 **4.91 s**; burst 200s p95 **2.91 s**. Full write-up: [`docs/load-test.md`](../load-test.md).

---

### Observability touchpoints

- **`LoggerWithID(r)`** — enriches logs with the request ID set by middleware (see `logging.md`)
- **`GET /metrics`** — Prometheus handler; basic-auth in production when configured, denied otherwise
- **`Panic recovery`** — middleware catches panics, logs stack, returns 500

Middleware details live in `internal/middleware/` — the API layer only consumes them in `NewRouter`.

---

### Process lifecycle (`cmd/server`)

The HTTP server is an explicit `http.Server` (not bare `ListenAndServe`):

| Concern | Behavior |
|---|---|
| Timeouts | `ReadHeaderTimeout` 5s, `ReadTimeout` 15s, `WriteTimeout` 60s, `IdleTimeout` 60s |
| Signals | `SIGINT` / `SIGTERM` via `signal.NotifyContext` |
| Shutdown | `Shutdown` with a 10s budget drains in-flight requests |
| Background | Stale-guest cleanup ticker stops when the signal context cancels |
| DB | `defer dbClient.Close()` runs after shutdown returns |

Render sends `SIGTERM` on deploy — without this path, in-flight requests are cut and the DB pool may not drain cleanly.

---

### Common mistakes to avoid

- Putting SQL or validation logic in handlers (old pattern — removed during refactor)
- Returning raw internal error strings on 500 responses
- Forgetting to pass `r.Context()` into service calls (breaks timeouts/cancellation)
- Adding admin routes without `RequireAdmin` wrapper
- Debugging "missing request ID in logs" without checking middleware order

