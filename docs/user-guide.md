# User Guide

How to use [infolinks.app](https://infolinks.app/) — for students and admins. For developers, see the [README](../README.md).

---

## Overview

**Info Links** helps CNAM Lebanon Computer Science students discover and organize course materials in one place — **50+ courses**, hundreds of curated links, and **300+ active users**.

- **Telegram updates:** [@Info_Links9](https://t.me/Info_Links9)
- **Contribute without code:** use **Report** or **Contribute** in the app nav

---

## For students

1. Open the site and select your **program** tab.
2. **Search** (`/` or `Ctrl+K`) or filter by year/semester to find a course.
3. Expand a course and open a link — badge color is the link type; label text is the content type (see legends below).
4. **Star** courses you revisit often (stored in your browser). Use **Report** or **Contribute** when a link is broken or you have a new resource.

Install from the browser menu as a **PWA** for a home-screen shortcut (service worker enabled in production builds).

### Student features

- **Smart search** — find courses by name or code (`/` or `Ctrl+K`)
- **Organized by program** — sorted by year, semester, and specialization
- **Favorites** — star courses for quick access (stored locally in your browser)
- **Content type labels** — TD, Cours, Videos, Sessions, Exams at a glance
- **Link type badges** — Google Drive, Classroom, Telegram, and more
- **Light/dark mode** — system detection with persistence
- **Report & contribute** — submit broken links or new resources
- **Feedback** — rate the platform (1–5 stars) by category
- **PWA** — installable with offline service worker support
- **SEO pages** — server-rendered course and program pages for search engines

---

## For admins

1. **Admin** → log in with your Supabase credentials (the API issues a JWT for the session).
2. **Courses** — manage courses and links; when prompted, confirm **sibling sync** for courses shared across programs.
3. **Contributions**, **Reports**, and **Feedback** — review and approve or resolve user submissions.
4. **Analytics** — check visitor counts and top clicked links; export JSON when needed.

### Admin features

- Full course and link CRUD with program/year/semester placement
- Optional vs. mandatory course labeling
- Sibling course detection — shared courses auto-sync names, codes, and links
- Multi-content link management (TD, Cours, Videos, Sessions, Exams)
- Analytics dashboard — daily visitors, 7/30/90-day ranges, top links, JSON export
- Contribution, report, and feedback review workflows
- JWT-secured admin panel via the Go API
- Extra resources sections beyond regular courses

---

## Content type legend

| Badge | Meaning |
|-------|---------|
| TD | Travaux Dirigés (exercises/tutorials) |
| Cours | Course materials/lectures |
| Videos | Video recordings |
| Exams | Exam papers and solutions |
| Other | Other types of content |

---

## Link type legend

| Badge | Meaning |
|-------|---------|
| **TG** | Telegram |
| **GD** | Google Drive |
| **GC** | Google Classroom |
| **OT** | Other / External |

---

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `/` or `Ctrl+K` | Focus search |
| `Esc` | Close modals |
