## SEO Layer in This Project

### What the `seo` package does

The `internal/seo` package serves **server-rendered HTML pages** for search engines and social crawlers.

The main SPA (React/Vite frontend) loads content via JavaScript — crawlers often see an empty shell. SEO routes return fully-formed HTML with:

- `<title>` and meta description
- Canonical URLs
- JSON-LD structured data (schema.org)
- Sitemap and robots.txt

This is separate from the JSON API in `internal/api`.

---

### Architecture

```text
seo.Handler (HTTP, HTML responses)
  → service.SEOService (thin pass-through)
  → repository.SEORepository (joined SQL queries)
  → seo.render / meta / jsonld / slug (pure functions)
```

Wired in `main`:

```go
seoHandler := seo.NewHandler(
    logger.With("component", "seo"),
    service.NewSEOService(repository.NewPostgresSEORepository(dbClient.DB)),
    cfg.SiteBaseURL,
)
```

Routes registered in `api.registerSEORoutes` — static file handler explicitly skips SEO paths.

---

### File roles

| File | Responsibility |
|---|---|
| `handlers.go` | HTTP handlers: course, program, index, sitemap, robots |
| `render.go` | HTML templates for pages (course, program, 404, 500, index) |
| `meta.go` | Title and description builders (~60 / ~155 char targets) |
| `jsonld.go` | schema.org JSON-LD for course pages |
| `slug.go` | `ProgramSlug` — accent folding, lowercase, hyphenate |
| `keywords.go` | Content type label helpers for meta text |

---

### Routes served

| Path | Handler | Response |
|---|---|---|
| `/course/{code}` | `HandleCourse` | Course page with links grouped by type |
| `/program/{slug}` | `HandleProgram` | Program page listing courses |
| `/courses` | `HandleCoursesIndex` | All courses index |
| `/sitemap.xml` | `HandleSitemap` | XML sitemap for crawlers |
| `/robots.txt` | `HandleRobots` | Allow/disallow rules + sitemap URL |

---

### Request flow (example: `/course/INF101`)

1. Extract `code` from path via `r.PathValue("code")`
2. `context.WithTimeout(r.Context(), 10s)` — cap DB/render time
3. `SEOService.GetCoursePageByCode(ctx, code)`
4. On `errs.ErrCourseNotFound` → render HTML 404 page
5. `renderCoursePage(baseURL, data)` → full HTML string
6. `writeHTML(w, 200, html)` with `Content-Type: text/html`

Errors during render → HTML 500 page with request ID for support.

---

### Rendering (`render.go`)

Uses Go's `html/template` with typed view structs:

- `coursePageView`, `programPageView`, `pageLayout`
- Links grouped into sections by content type for display
- Shared layout wraps title, description, canonical, JSON-LD, body

All user-facing strings escaped via template — prevents XSS in rendered pages.

---

### Meta tags (`meta.go`)

`BuildCourseTitle` and `BuildCourseDescription` produce SEO-friendly strings:

- Include course name, code, content types (TD, examens, …)
- Truncate if over ~65 / ~160 chars
- French copy targeting CNAM Liban search terms

Derived from repository data (`CoursePageData`, link content types).

---

### JSON-LD (`jsonld.go`)

`buildCourseJSONLD` emits schema.org graph:

- `@type: Course` with name, courseCode, provider
- `@type: ItemList` of link resources

 Helps Google understand page content beyond plain HTML.

---

### Slugs (`slug.go`)

`ProgramSlug(name)` converts program names to URL-safe slugs:

1. Trim, lowercase
2. Fold accents (`é` → `e`) via `strings.NewReplacer`
3. Replace non-alphanumeric runs with `-`
4. Trim trailing hyphens

Used when generating program URLs and when matching `/program/{slug}` requests (repo receives slug fn as parameter).

---

### Sitemap (`HandleSitemap`)

Builds XML manually with `strings.Builder`:

- Home page, `/courses`
- Every course code → `/course/{code}`
- Every program → `/program/{slug}`

XML special characters escaped via `xmlEscape`. 20-second context timeout for large DB.

---

### Robots.txt

```
User-agent: *
Allow: /
Disallow: /admin
Disallow: /admin-gate
Sitemap: {baseURL}/sitemap.xml
```

Keeps admin UI out of crawl index while allowing public course pages.

---

### Interaction with static file handler

`router.go` static handler checks `isSEOPath` — SEO routes never fall through to SPA `index.html`.

SPA paths (`/`, `/admin`, …) get `index.html`; SEO paths always hit `seo.Handler`.

---

### Timeouts

SEO handlers set explicit context timeouts (10–20s) because page rendering may run heavier queries than simple API calls. Prevents slow DB from hanging connections indefinitely.

---

### Quick decision guide

- Crawler-facing HTML → `internal/seo`
- JSON for frontend SPA → `internal/api`
- SEO-specific SQL joins → `repository/seo.go` + `seo_models.go`
- URL slug rules → `seo/slug.go`
- Title/description copy → `seo/meta.go`

---

### Common mistakes to avoid

- Serving SEO content only client-side — crawlers won't execute JS reliably
- Forgetting to exclude SEO paths from SPA fallback
- Building HTML with string concat and user data — use templates + escape
- Hardcoding base URL — use `cfg.SiteBaseURL` passed to handler

