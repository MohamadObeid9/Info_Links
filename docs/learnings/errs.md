## Errors in This Project

### What the `errs` package does

The `internal/errs` package defines **sentinel errors** — fixed, comparable error values used across service, repository, and handler layers.

```go
var ErrLinkNotFound = errors.New("link not found")
```

Handlers use `errors.Is(err, errs.ErrLinkNotFound)` to pick HTTP status codes without string matching.

Domains covered today include links, courses, reports, feedback, contributions, users, extra sections/links, and community services (`ErrServiceNotFound`, `ErrServiceInvalidStatus`, `ErrServiceInvalidRenewal`, …).

---

### Why sentinel errors exist

Go errors are opaque strings by default. Sentinels give you:

- **Stable identity** — compare with `errors.Is`, not `err.Error() == "..."`
- **Cross-layer contracts** — repository returns `ErrLinkNotFound`, handler maps to 404
- **Testability** — assert exact error type in service/repo tests

Two error categories in this project:

| Type | Example | Used for |
|---|---|---|
| Sentinel (`errors.New`) | `ErrLinkNotFound` | Expected domain outcomes |
| Wrapped (`fmt.Errorf %w`) | `fmt.Errorf("delete link: %w", err)` | Unexpected infrastructure failures |

See `logging.md` for when to use each.

---

### Error flow through layers

Example: delete link with invalid ID

```text
Handler:  idStr from path → linkService.Delete(ctx, idStr)
Service:  strconv.Atoi fails → return errs.ErrLinkInvalidID
Handler:  mapDeleteLinkErr → errors.Is(err, ErrLinkInvalidID) → 400
```

Example: delete link that does not exist

```text
Repository: DELETE ... RowsAffected == 0 → return errs.ErrLinkNotFound
Service:    wraps or passes through
Handler:    mapDeleteLinkErr → 404
```

Example: database connection failure

```text
Repository: ExecContext fails → fmt.Errorf("delete link: %w", err)
Service:    fmt.Errorf("delete link: %w", err)
Handler:    default case → log + 500
```

---

### Who creates which errors

| Layer | Creates |
|---|---|
| Service | Validation errors (empty fields, bad enums, invalid params) |
| Repository | Not-found after write operations (`RowsAffected == 0`) |
| Handler | Never creates domain errors — only maps them to HTTP |

---

### Handler mapping pattern

Each handler file has `mapXxxErr` helpers:

```go
switch {
case errors.Is(err, errs.ErrLinkNotFound):
    writeJSONError(w, r, http.StatusNotFound, "Link not found")
case errors.Is(err, errs.ErrLinkInvalidID):
    writeJSONError(w, r, http.StatusBadRequest, "Invalid link id")
default:
    h.LoggerWithID(r).Error("delete link failed", "error", err)
    writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
}
```

**Rules:**

- Sentinel → specific 4xx + safe user message
- Wrapped/unknown → log + generic 500
- Never expose internal error strings to clients
