## Database Layer in This Project

### What the `database` package does

The `internal/database` package owns **Postgres connection setup** — opening the pool, configuring limits, ping on startup, and graceful close.

It does not run queries. All SQL lives in `internal/repository`.

---

### Core type: `Client`

```go
type Client struct {
    DB *sql.DB
}
```

`database.Client` wraps `*sql.DB` and adds:

- `Ping(ctx)` — used by `/readyz` readiness checks
- `Close()` — called from `main` defer on shutdown

The handler depends on a small `dbPinger` interface (only `Ping`), not the full client — keeps health checks decoupled.

---

### Constructor: `New(dbUrl, logger)`

```go
func New(dbUrl string, logger *slog.Logger) (*Client, error)
```

Steps:

1. Validate logger and `dbUrl` are non-empty
2. `sql.Open("pgx", dbUrl)` — pgx driver registered via blank import
3. Configure connection pool
4. `PingContext` with 5-second timeout — fail fast if DB unreachable
5. Log success, return `Client`

If ping fails, closes the DB and returns wrapped error — `main` exits.

---

### Connection pool settings

```go
db.SetMaxOpenConns(20)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(30 * time.Minute)
db.SetConnMaxIdleTime(10 * time.Minute)
```

| Setting | Value | Why |
|---|---|---|
| MaxOpenConns | 20 | Cap concurrent connections to Postgres |
| MaxIdleConns | 10 | Keep warm connections without hoarding |
| ConnMaxLifetime | 30m | Recycle connections (useful behind PgBouncer / Supabase) |
| ConnMaxIdleTime | 10m | Drop unused idle connections |

These are reasonable defaults for a single-instance app on Render talking to Supabase Postgres.

---

### Driver choice: pgx via `database/sql`

```go
import _ "github.com/jackc/pgx/v5/stdlib"
```

Uses the standard library `database/sql` interface with the pgx v5 driver. Repositories work with `*sql.DB` — no pgx-specific API in the repo layer.

Benefits:

- Familiar `QueryContext` / `ExecContext` patterns
- Easy `sqlmock` testing
- Can swap drivers without changing repository code

---

### Wiring in `main`

```go
dbClient, err := database.New(cfg.DatabaseURL, logger.With("component", "database"))
if err != nil {
    logger.Error("database initialization failed", "error", err)
    os.Exit(1)
}
defer func() { _ = dbClient.Close() }()

services := handleServices(dbClient.DB)
```

- Config provides `DatabaseURL` (from env / Supabase)
- Logger tagged with `component=database`
- `dbClient.DB` passed to all repository constructors
- `dbClient` passed to API handler for readiness ping

---

### No global singleton

The old pattern was `var DB *sql.DB` — convenient but untestable.

Now:

- One `Client` created in `main`
- Passed explicitly to repos and handler
- No package-level mutable database state

---

### Health vs readiness

| Endpoint | Uses database package? |
|---|---|
| `GET /healthz` | No — always OK if process runs |
| `GET /readyz` | Yes — `h.db.Ping(ctx)` with 5s timeout |

Render's health check hits `/readyz`. If Supabase is down, the instance is marked not ready and traffic stops.

---

### Quick decision guide

- Open/close/pool config → `database` package
- SQL queries → `repository` package
- "Is DB up?" for load balancer → `/readyz` via handler's `dbPinger`
- Connection string → `config.DatabaseURL`, never hardcoded

