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
4. **Star** courses you revisit often. Use **Report** or **Contribute** when a link is broken or you have a new resource.

Install from the browser menu as a **PWA** for a home-screen shortcut (service worker enabled in production builds).

### Signing up

There is **no email and no password**. You identify yourself with three things:

- your **first name**
- your **last name**
- a **number between 1 and 100** you pick yourself

That becomes your handle — for example `mohamad_hassan_55` — and the home page greets you with it once you are signed in.

- **Sign up** the first time. If someone with your exact name already picked your number, the app asks you to choose a different one (55 becomes 65).
- **Sign in** on any other device with the same three values. If we cannot find them, the app points you to sign up instead.
- Your session lasts a **year** on that device, so you normally type this once.

Keep your number in mind — name plus number is how the app recognizes you, and there is no password to reset if you forget which number you chose.

### What needs an account

Browsing never does. Search, filter, and look through courses without signing up at all.

| Action | Account needed |
|--------|----------------|
| Browse, search, and filter courses | No |
| Open a course link | Yes |
| Star a course (favorites) | Yes |
| Report a broken link | Yes |
| Contribute a new resource | Yes |
| Send feedback | Yes |

When you try one of these without an account, a short signup form appears and the action continues right after you finish.

### Student features

- **Smart search** — find courses by name or code (`/` or `Ctrl+K`)
- **Organized by program** — sorted by year, semester, and specialization
- **Account without a password** — first name, last name, and a number 1-100
- **Favorites that follow you** — starred courses are saved to your account and appear on every device you sign in on
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
2. **Courses** — manage offerings and links. The same course code is one catalog row; adding it to another program shares its links automatically. Deleting a course from one program leaves it in others.
3. **Contributions**, **Reports**, and **Feedback** — review user submissions. Contributions can be **approved** (adds the link) or **rejected** (kept in the list, not deleted). Each row shows the **sender's handle**.
4. **Analytics** — unique students per range alongside visit counts and top clicked links; export JSON when needed.
5. **Students** — browse every registered student, search by name, and open one to see their full history.

Admin login is unchanged and stays separate from student accounts: admins sign in with Supabase credentials, students never do.

### Students tab

- Alphabetical list of registered students with first seen and last seen dates
- Search by handle
- Detail view with a single activity timeline — visits (e.g. “visited home from phone”), links opened, reports, contributions, feedback, and favorites added or removed, newest first. The profile card shows the device from their most recent classified visit.

Activity from before a student signed up is kept and appears in their timeline, because the visitor record is claimed at signup rather than replaced. If they already have an account and sign in instead, that first guest visit is moved onto their student row so it shows their handle, not `guest_<id>`.

### Unique-user analytics

The Analytics tab counts **people**, not just page loads:

- **Overview cards** — registered students, active/clicks/device **in the selected 7/30/90 range** (with vs-prior deltas), unique phone/laptop students, returning vs new, signup funnel, browse depth, and open inbox counts
- **Growth chart** — toggle between unique visitors per day and cumulative registered roster over the selected range
- **Today's visitors** — paged handle chips (sort by clicks or name)
- **Demand** — top links, top courses, zero-click courses/links, most-favorited courses, top students
- **Search terms** and an hour×weekday activity heatmap

New visits store a coarse `device_type` (`phone` or `laptop`) derived from the User-Agent on the server; the client never sends device data. Search and browse-depth events are recorded separately so demand and funnel metrics stay actionable.

Counting is done in the database rather than in the browser, so the dashboard stays fast as history grows. Two caveats when reading the numbers: visitors who never sign up are still counted as visits but cannot be attributed to a person, and admin browsing is excluded from analytics entirely.

### Admin features

- Full course and link CRUD with program/year/semester placement
- Optional vs. mandatory course labeling
- Sibling course detection — shared courses auto-sync names, codes, and links
- Multi-content link management (TD, Cours, Videos, Sessions, Exams)
- Analytics dashboard — overview cards, growth chart, paged today's visitors, top links/students, JSON export
- Students directory with per-student activity timeline
- Sender handle on contributions, reports, and feedback
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
