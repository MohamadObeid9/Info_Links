## Repository Layer in This Project

### What the `repository` package does

The `internal/repository` package is the **database access layer**. It owns:

- SQL queries (as constants in `queries.go`)
- Postgres implementations behind interfaces
- Row scanning into models or DTOs
- Translating "zero rows affected" into sentinel errors

Repositories know **nothing about HTTP** and **nothing about business validation** beyond what SQL requires.

---

### Core pattern

```text
Interface (repositories.go)
  → postgresXxxRepository struct { db *sql.DB }
  → NewPostgresXxxRepository(db) returns the interface
  → methods use QueryContext / ExecContext with ctx
```

Example:

```go
type postgresLinkRepository struct {
    db *sql.DB
}

func NewPostgresLinkRepository(db *sql.DB) LinkRepository {
    return &postgresLinkRepository{db: db}
}
```

Services depend on interfaces (`LinkRepository`), not the concrete struct — tests swap in mocks.

---

### File roles

| File | Responsibility |
|---|---|
| `repositories.go` | All repository interfaces |
| `queries.go` | SQL query constants grouped by domain |
| `links.go`, `courses.go`, … | Postgres implementation per domain |
| `helpers.go` | Shared list-query builder for filtered admin lists |
| `seo.go`, `seo_models.go` | SEO-specific queries and DTOs (not `models` package) |

---

### Repository interfaces

Defined in `repositories.go`:

- `LinkRepository`, `CourseRepository`, `ReportRepository`, …
- `ServiceRepository` — community listings + click tracking + freeze-expired
- `ContentRepository` — returns pre-built JSON bytes
- `SEORepository` — returns SEO page DTOs (`CoursePageData`, etc.)

Each interface is small — only the methods that service layer needs.

---

### SQL organization (`queries.go`)

All SQL lives as named constants:

```go
const (
    insertLinkQuery = `INSERT INTO links (...) VALUES ($1, $2, ...)`
    deleteLinkQuery = `DELETE FROM links WHERE id = $1`
)
```

**Why constants in one file:**

- Easy to review all queries in one place
- Tests can reference the same query string for `sqlmock`
- No string literals scattered through methods

Parameterized queries always use `$1, $2, …` — never string concatenation for user input.

---

### Create / update / delete patterns

**Insert:** `ExecContext` — check error only.

**Update / delete:** `ExecContext` + `RowsAffected`:

```go
affected, err := resp.RowsAffected()
if affected == 0 {
    return errs.ErrLinkNotFound
}
```

Zero rows means the ID did not exist — return a sentinel, not a generic error.

---

### Filtered list queries (`helpers.go`)

Admin list endpoints (reports, feedback, contributions) support optional `q` (search) and `status` filters.

`buildFilteredListQuery` picks the right SQL variant:

| Filters | Query used |
|---|---|
| none | `noFilter` |
| `q` only | `withQ` |
| `status` only | `withStatus` |
| both | `withQStatus` |

Each domain defines four query constants; the helper returns `(query, args)`.

---

### Content repository — one big JSON query

`ContentRepository.Get` runs a single Postgres query that builds the entire navigation tree as JSON:

- Subqueries with `json_agg` for each table (years, courses, programs, …)
- `json_build_object` wraps everything into one payload
- Returns `[]byte` — handler writes it directly without re-encoding

This avoids N+1 queries and keeps the frontend's one-shot `/api/content` fetch fast.

---

### SEO repository

SEO has its own DTOs in `seo_models.go` (`CoursePageData`, `SEOLink`, `CoursePlacement`, …) because page rendering needs joined data, not raw `models.Course` structs.

`postgresSEORepository`:

- Joins courses → semesters → years → programs for placement info
- Fetches links for course IDs in a second query
- Returns `errs.ErrCourseNotFound` when no rows match a course code

Slug generation (`ProgramSlug`) is passed in as a function parameter where needed — keeps URL logic in `internal/seo`, not SQL.

---

### Context and errors

Every method signature: `func (r *...) Method(ctx context.Context, ...) error`

- Always pass `ctx` to `QueryContext`, `QueryRowContext`, `ExecContext`
- Wrap driver errors: `fmt.Errorf("delete link: %w", err)`
- Return sentinels for domain outcomes: `errs.ErrLinkNotFound`, `errs.ErrCourseNotFound`

Do not import `internal/api` or return HTTP concepts.

---

### Testing

Repository tests use `sqlmock` to assert:

- The correct query constant is executed
- Args match expected values
- Scan/error paths behave correctly

See `links_test.go`, `reports_test.go`, etc.
