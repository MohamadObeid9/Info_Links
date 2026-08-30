## Service Layer in This Project

### What the `service` package does

The `internal/service` package holds **business logic** between HTTP handlers and the database.

Each service:

- Validates and normalizes input (trim strings, parse IDs, check enums)
- Returns typed sentinel errors from `internal/errs`
- Delegates persistence to a repository interface
- Wraps unexpected failures with `fmt.Errorf("...: %w", err)`

Services know **nothing about HTTP** — no status codes, no JSON, no headers.

---

### Why a service layer exists

Before the refactor, SQL and validation lived in handlers. That made handlers fat and hard to test.

Now the split is:

```text
Handler  → decode HTTP, map errors to status codes
Service  → validate, apply rules, orchestrate
Repository → run SQL
```

**Benefits:**

- Test business rules without `httptest` or a real database (mock the repo)
- Reuse logic if you add a CLI or worker later
- Handlers stay thin (~10 lines per method)

---

### One service per domain

| Service | File | Repository |
|---|---|---|
| `LinkService` | `links.go` | `LinkRepository` |
| `CourseService` | `courses.go` | `CourseRepository` |
| `ReportService` | `reports.go` | `ReportRepository` |
| `FeedbackService` | `feedbacks.go` | `FeedbackRepository` |
| `ContributionService` | `contributions.go` | `ContributionRepository` |
| `ContentService` | `contents.go` | `ContentRepository` |
| `PageViewService` | `pageviews.go` | `PageViewRepository` |
| `LinkClickService` | `linkclicks.go` | `LinkClickRepository` |
| `ExtraSectionService` | `extra_sections.go` | `ExtraSectionRepository` |
| `ExtraLinkService` | `extra_links.go` | `ExtraLinkRepository` |
| `ServiceService` | `services.go` | `ServiceRepository` |
| `UserService` | `users.go` | `UserRepository` |
| `AnalyticsService` | `analytics.go` | `AnalyticsRepository` |
| `SEOService` | `seo.go` | `SEORepository` |

All follow the same constructor pattern:

```go
type LinkService struct {
    repo repository.LinkRepository
}

func NewLinkService(repo repository.LinkRepository) *LinkService {
    return &LinkService{repo: repo}
}
```

Wiring happens in `cmd/server/main.go` inside `handleServices`.

---

### Typical method flow (example: `LinkService.Create`)

1. **Normalize** — `strings.TrimSpace` on URL and label
2. **Validate** — return `errs.ErrLinkURLAndLabelRequired` if empty
3. **Persist** — `s.repo.Create(ctx, link)`
4. **Wrap errors** — `fmt.Errorf("create link: %w", err)` for unexpected DB failures

The handler decides: `ErrLinkURLAndLabelRequired` → 400, wrapped DB error → 500.

---

### ID parsing pattern

Path params arrive as strings from HTTP. Services parse them:

```go
idStr = strings.TrimSpace(idStr)
id, err := strconv.Atoi(idStr)
if err != nil || id <= 0 {
    return errs.ErrLinkInvalidID
}
```

Repositories receive `int` IDs — parsing is a service concern, not SQL.

---

### Partial updates (`CourseService.Update`)

Courses support PATCH with optional fields via `models.CoursePatch` (pointer fields):

1. Parse and validate ID
2. Load existing row with `repo.GetByID`
3. Reject empty patch (`ErrCoursePatchEmpty`)
4. Merge non-nil patch fields into existing course
5. Re-validate merged result (name, code, semester)
6. Call `repo.Update`

This keeps PATCH logic out of the handler and repository.

---

### List + filter validation (`ReportService.List`)

List endpoints validate pagination and filters in the service:

- `limit` must be 1–100, `offset` ≥ 0 → else `ErrInvalidParams`
- `status` must be `"open"`, `"resolved"`, or empty → else `ErrReportInvalidStatus`

The repository receives clean, validated values and builds the right SQL query.

---

### Community services (`ServiceService`)

Community listings (tutoring, student businesses) use status `trial` / `active` / `frozen`:

- **Create** — defaults to a 15-day trial when `expires_at` is omitted
- **List** — freezes expired rows before returning results
- **Renew / freeze / unfreeze** — admin lifecycle actions with validated day ranges (1–365)
- **TrackClick** — records student opens via `POST /api/service_clicks`

Same validation → sentinel → repository pattern as courses and links.

---

### Thin services vs rich services

Some services are thin pass-throughs (`ContentService.Get`, `PageViewService.Create`) — the repo does the heavy lifting (e.g. one big JSON aggregation query).

Others are rich (`CourseService.Update`, `ReportService.List`) — validation and orchestration live here.

Both are fine. Add logic to the service when it is not SQL-specific and not HTTP-specific.

---

### Context usage

Every service method takes `context.Context` as the first argument and passes it to the repository.

This enables:

- Request cancellation when the client disconnects
- Timeouts set by callers (SEO handlers use `context.WithTimeout`)
- Consistent propagation through the stack

Never use `context.Background()` inside a service when you received a context from above.

---

### Error handling rules

| Situation | Return |
|---|---|
| Invalid input the client can fix | Sentinel from `internal/errs` |
| Row not found after delete/update | Sentinel (e.g. `ErrLinkNotFound`) — set by repository |
| Database/driver failure | `fmt.Errorf("operation: %w", err)` |

Services do **not** log. Handlers log on 500; services just return errors.
