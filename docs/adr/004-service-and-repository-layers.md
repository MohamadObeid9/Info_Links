# ADR 004: Service and Repository Layers

## Status

Accepted (refactored from monolithic handlers, 2026)

## Context

The original backend had ~6 files with SQL and validation embedded in HTTP handlers and a global `var DB *sql.DB` singleton. That worked for a fast MVP but caused:

- Handlers too large to test without HTTP and a real database
- Business rules mixed with status codes and JSON encoding
- Impossible to mock the database in unit tests
- Risk in interviews: "walk me through your code" exposed shallow understanding

We needed a structure that survives a senior engineer's review and supports ≥70% test coverage on API and service packages.

Alternatives considered:

- **Handlers-only + sqlc** — SQL in `.sql` files, generated Go; still need a place for validation
- **Single "store" package** — simpler, but mixes validation and SQL
- **Handler → Repository (skip service)** — fewer layers, but validation ends up in handlers or repos
- **Handler → Service → Repository** — classic layered architecture with interfaces for testing

## Decision

Introduce two internal layers:

**`internal/repository`**

- Interfaces per domain (`LinkRepository`, `CourseRepository`, …)
- Postgres implementations with SQL in `queries.go`
- Returns sentinel errors (`errs.ErrLinkNotFound`) on zero-row updates/deletes

**`internal/service`**

- One service per domain, depends on repository interfaces
- Input validation, ID parsing, enum checks, partial-update merging
- No HTTP, no logging — returns errors upward

**`internal/api`**

- Thin handlers: decode JSON, call service, map errors to status codes
- Depends on small service interfaces defined in `handler.go` for mocking

**Dependency injection in `main`:**

- `database.New()` returns a client (no global singleton)
- `handleServices(db)` wires repo → service
- `api.NewHandler(deps)` receives all services explicitly

## Consequences

### Positive

- `internal/api` at ~93% and `internal/service` at ~96% test coverage with table-driven tests and mocks
- Each layer has a single reason to change
- Interview story: clear request flow from router → handler → service → repo → Postgres
- Companion worker or CLI can reuse services without duplicating validation

### Negative

- More files and interfaces for a small app — some services are thin pass-throughs
- Wiring in `main` is repetitive (10 repo/service pairs)
- Must maintain parallel interfaces in `handler.go` and `repositories.go`
- Refactor cost already paid; new domains still require three touch points (handler, service, repo)

## References

- `cmd/server/main.go` — `handleServices`, `NewHandler`
- `internal/api/handler.go` — service interfaces
- `internal/repository/repositories.go` — repository interfaces
- `docs/learnings/service.md`, `docs/learnings/repository.md`, `docs/learnings/errs.md`
