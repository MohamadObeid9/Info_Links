# ADR 004: Service and Repository Layers

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

- `internal/api` at ~79% and `internal/service` at ~73% statement coverage with table-driven tests and mocks (~74% across all packages)
- Each layer has a single reason to change
- Interview story: clear request flow from router → handler → service → repo → Postgres
- Companion worker or CLI can reuse services without duplicating validation

