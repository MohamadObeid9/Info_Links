## Middleware in This Project

### What the `middleware` package does

The `internal/middleware` package provides **HTTP middleware** — functions that wrap handlers to add cross-cutting behavior before or after request processing.

Middleware runs on every request (unless exempted) and handles concerns that should not be duplicated in each handler:

- Request IDs and access logging
- Prometheus metrics
- Rate limiting
- Panic recovery
- Admin JWT auth (per-route, not global)

---

### Where middleware is applied

**Global chain** — built in `api.NewRouter`:

```text
Request (outermost first)
  → CORS
  → RequestIDWithLogging
  → Metrics
  → RateLimit
  → Recover
  → Security headers
  → ServeMux (routes)
```

**Per-route** — `middleware.RequireAdmin` wraps individual admin handlers in `registerAdminRoutes`. `middleware.RequireUser` wraps student-authenticated routes (page views, link clicks, service clicks, favorites, etc.).

**Per-endpoint** — `MetricsBasicAuth` / `MetricsDenied` wrap the Prometheus handler only.

---

### File roles

| File | Middleware | Purpose |
|---|---|---|
| `requestid.go` | `RequestID`, `RequestIDWithLogging` | Assign/propagate `X-Request-ID`, access logs |
| `metrics.go` | `Metrics` | Prometheus counters and latency histograms |
| `metrics_auth.go` | `MetricsBasicAuth`, `MetricsDenied` | Protect `/metrics` in production |
| `ratelimit.go` | `RateLimit` | Per-IP token bucket rate limiting |
| `recover.go` | `Recover` | Catch panics, log stack, return 500 JSON |
| `auth_middleware.go` | `RequireAdmin` | JWT validation for admin routes |

---

### Request ID (`requestid.go`)

Every request gets an `X-Request-ID`:

- Reuses client-provided ID if valid 
- Otherwise generates 32-char hex from `crypto/rand`
- Stores ID in `context.Context` via unexported `contextKey`
- Sets response header `X-Request-ID`

`RequestIDFromContext(ctx)` is used by handlers and `Recover` to correlate logs.

`RequestIDWithLogging` also logs one access line per request:

```text
level=INFO msg="request completed" request_id=... method=GET path=/api/content status=200 duration_ms=12
```

Skips noise: static assets, `/readyz`, HEAD requests, `/metrics` in production.

---

### Metrics (`metrics.go`)

Prometheus metrics registered with `promauto`:

- `http_requests_total{method, path, status}` — counter
- `http_request_duration_seconds{method, path}` — histogram

Uses a wrapping `responseWriter` to capture the final status code.

**Path normalization** — dynamic segments collapsed to reduce cardinality:

- `/api/admin/links/42` → `/api/admin/links/{id}`
- `/course/INF101` → `/course/{code}`
- `/program/genie-informatique` → `/program/{slug}`

Skips: `/metrics`, `/healthz`, `/readyz`, static assets (`.js`, `.css`, images).

---

### Metrics auth (`metrics_auth.go`)

Configured in `router.go` via `metricsHandler(cfg)`:

| Config | Behavior |
|---|---|
| Basic auth credentials set | `MetricsBasicAuth` protects endpoint |
| Production, no credentials | `MetricsDenied` — always 401 |
| Development | Open Prometheus handler |

Uses `crypto/subtle.ConstantTimeCompare` to avoid timing attacks on credentials.

---

### Rate limiting (`ratelimit.go`)

Per-IP token bucket using `golang.org/x/time/rate`:

- **10 req/s** sustained, **20 burst**
- In-memory map of IP → limiter, protected by `sync.Mutex`
- Background goroutine evicts idle limiters after 10 minutes

**Client IP detection:**

- Default: `r.RemoteAddr`
- If remote is a trusted proxy (loopback, private ranges), parse `X-Forwarded-For` right-to-left for first non-trusted IP

**Exempt paths:** `/healthz`, `/readyz`, `/metrics`, `/robots.txt`, static assets (paths with file extensions).

Over limit → `429` with `{"error":"rate limit exceeded"}`.

---

### Panic recovery (`recover.go`)

`defer recover()` around every request:

- Logs panic value, method, path, full stack trace
- Includes `request_id` when available
- Returns JSON 500 with generic message + request ID
- Process keeps running — one bad request does not crash the server

---

### Admin auth (`auth_middleware.go`)

`RequireAdmin(jwtSecret, next)` — not global; applied per admin route.

Steps:

1. Read `Authorization` header
2. Strip `Bearer ` prefix if present
3. Parse JWT with HMAC-SHA256, verify signature with `jwtSecret`
4. Check `admin: true` claim in `MapClaims`
5. Call `next` or return 401/403 JSON

This validates the **app JWT** issued by `POST /api/auth/login`, not the Supabase token.

---

### Middleware order matters

Built inside-out in `NewRouter`:

```go
handlerWithRecover := middleware.Recover(logger, securedHandler)
handlerWithRateLimit := middleware.RateLimit(handlerWithRecover)
handlerWithMetrics := middleware.Metrics(handlerWithRateLimit)
handlerWithRequestID := middleware.RequestIDWithLogging(logger, cfg.AppEnv, handlerWithMetrics)
```

**Why this order:**

- **Recover** innermost — catches panics from handlers and other middleware
- **RateLimit** before metrics — rejected requests still counted if placed after metrics (current: rate limit runs before handler, metrics wraps everything including rate limit response)
- **RequestID** outermost (before CORS) — ID available for all inner layers and access log captures full duration

If debugging "no request ID in panic log", check that RequestID runs before Recover in the chain (it does — RequestID is outer).

---

### Quick decision guide

- Cross-cutting, every request → global middleware in `NewRouter`
- Auth on specific routes → per-route wrapper like `RequireAdmin`
- Need value in handlers → store in context, expose getter (`RequestIDFromContext`)
- Protect one endpoint differently → wrap that handler only (metrics auth)
