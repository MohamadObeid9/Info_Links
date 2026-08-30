## Models in This Project

### What the `models` package does

The `internal/models` package defines **shared domain structs** — the shape of data as it moves between handlers, services, and repositories.

It contains:

- Struct definitions with JSON tags
- No methods, no validation, no SQL
- No imports from other internal packages

Think of it as the vocabulary of the domain.

---

### Entity hierarchy (navigation tree)

The app's core data model is a tree:

```text
Program
  └── Year
        └── Semester
              └── Course
                    └── Link
```

Plus parallel structures:

```text
ExtraSection
  └── ExtraLink

Service  (community listing; links stored as JSONB)
ServiceClick
```

**Structs:** `Program`, `Year`, `Semester`, `Course`, `Link`, `ExtraSection`, `ExtraLink`, `Service`, `ServiceLink`, `ServiceClick`

Each has `DisplayOrder` for frontend sorting where applicable. Foreign keys use `ProgramID`, `YearID`, `SemesterID`, `CourseID`, `SectionID`, `ServiceID`.

---

### Analytics entities

| Struct | Purpose |
|---|---|
| `PageView` | Tracks which page was visited |
| `LinkClick` | Tracks clicks on a specific link ID |
| `ServiceClick` | Tracks opens on a community service card |
| `ServiceDemand` | Aggregated service open counts for analytics |

Simple structs — `PageView` has `Page`, `LinkClick` has `LinkID`, `ServiceClick` has `ServiceID`.

---

### Partial update: `CoursePatch` and `ServicePatch`

```go
type CoursePatch struct {
    Name       *string `json:"name"`
    Code       *string `json:"code"`
    IsOptional *bool   `json:"is_optional"`
    SemesterID *int    `json:"semester_id"`
}
```

Pointer fields mean "only update if present in JSON":

- `nil` → leave existing value
- non-nil → apply new value

Used by `PATCH /api/admin/courses/{id}` and `PATCH /api/admin/services/{id}` (`ServicePatch`). Merging happens in the service layer, not in the model.

---

### JSON tags

All exported fields have `` `json:"snake_case"` `` tags matching API conventions:

```go
type Link struct {
    ID           int     `json:"id"`
    URL          string  `json:"url"`
    Label        string  `json:"label"`
    CourseID     *int    `json:"course_id,omitempty"`
    ContentType  *string `json:"content_type"`
}
```

- `omitempty` on optional pointers — omitted from JSON when nil
- Handlers decode request bodies directly into these structs

---

### `ContentResponse`

Aggregates the full navigation payload:

```go
type ContentResponse struct {
    Links         []Link
    Years         []Year
    Courses       []Course
    Programs      []Program
    Semesters     []Semester
    ExtraLinks    []ExtraLink
    ExtraSections []ExtraSection
    Services      []Service
}
```

The live `/api/content` endpoint returns JSON built by Postgres (`json_build_object`), not by marshaling this struct in Go. Community services are also available via `GET /api/services`. The struct documents the expected shape.
