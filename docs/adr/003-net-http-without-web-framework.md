# ADR 003: `net/http` Without a Web Framework

## Status

Accepted

## Context

The Go backend exposes REST JSON endpoints, serves static frontend assets, and renders SEO HTML pages. We needed a routing and middleware story that:

- Keeps dependencies minimal
- Demonstrates understanding of Go's HTTP stack (relevant for backend interviews)
- Supports method-based routes (`GET /api/content`, `PATCH /api/admin/links/{id}`)
- Allows a custom middleware chain (metrics, rate limit, request ID, panic recovery)

Popular alternatives:

- **Gin, Echo, Fiber** — fast to scaffold, large ecosystems, familiar to many juniors
- **Chi** — lightweight router on `net/http`, still an extra dependency
- **Standard library `net/http` ServeMux** — Go 1.22+ supports method and path patterns natively

## Decision

Build the HTTP layer on **Go standard library only**:

- `http.NewServeMux()` with Go 1.22+ route patterns (`"GET /api/content"`, `"PATCH /api/admin/links/{id}"`)
- Custom middleware as `func(http.Handler) http.Handler` wrappers in `internal/middleware`
- Small third-party additions only where stdlib is weak: **CORS** (`rs/cors`), **Prometheus** (`client_golang`), **JWT** (`golang-jwt/jwt`)

No Gin, Echo, Fiber, or Chi.

## Consequences

### Positive

- Senior Go engineers often prefer stdlib-heavy code — shows deliberate choice, not tutorial default
- Full visibility into middleware order and request flow — nothing hidden in framework magic
- Fewer dependencies to audit and upgrade
- Go 1.22 ServeMux path params (`r.PathValue("id")`) cover admin CRUD needs

### Negative

- More boilerplate than Gin for grouping routes or binding JSON to structs
- No built-in request validation framework — manual `decodeJSONBody` and service validation
- Middleware composition is manual — must document order (see `docs/learnings/middleware.md`)
- Less copy-paste from Stack Overflow answers that assume Gin/Fiber

## References

- `internal/api/router.go` — route registration and middleware chain
- `internal/api/helpers_http.go` — JSON helpers
- `docs/learnings/api.md`
- `docs/learnings/middleware.md`
