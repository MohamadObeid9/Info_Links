# ADR 009: Canonical Courses and Placements

## Status

Accepted

## Context

The same CNAM course (same code, e.g. GDN100) is offered in Licence, AISL, and IRSM. The original schema stored **one `courses` row per program**, each with its own `links`. Admins then copied links across “siblings,” analytics listed the same course three times, and clicks on identical URLs were split across duplicate `link_id`s.

We needed one identity for a course and one set of resources, while still showing that course in every program that offers it.

## Decision

1. **`courses`** is the canonical catalog row: `name`, `code`, `is_optional`. Codes are unique (case-insensitive) when non-empty.
2. **`course_placements`** is an offering: `course_id` + `semester_id` + `display_order`. The SPA tree still nests courses under semesters; `/api/content` expands each placement into a course object with `placement_id` and `semester_id`.
3. **`links`** belong to the canonical course, not to a placement. Adding or editing a link once updates every program that offers that course.
4. Existing duplicate rows (same `lower(trim(code))`) are merged in migration `000009`. Duplicate links (same course + URL) are merged and `link_clicks` are remapped to the surviving `link_id`.
5. Favorites stay `course_id` arrays and therefore apply to the course everywhere it appears.

Per-program click splits are no longer stored as separate link copies. Historical clicks from merged URLs are kept on the surviving link. New analytics can still list which programs *offer* a course via placements.

## Consequences

- Admin no longer asks “apply to all siblings?” for links or name/code edits
- Removing a course from one program deletes that placement only; the last placement deletes the course and its links
- Adding a course whose code already exists attaches a new placement instead of cloning the catalog
- SEO pages that already grouped by code become a single course with multiple placements
- Restoring an old admin backup must collapse duplicate codes again (seed does this)
- Down-migration of `000009` cannot recreate the pre-merge copies; it is not lossless
