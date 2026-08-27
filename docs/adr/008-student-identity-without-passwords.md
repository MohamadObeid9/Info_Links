# ADR 008: Student Identity Without Passwords

## Context

Info Links could not answer a basic question: **how many people actually use it?** Analytics stored anonymous rows — `page_views` and `link_clicks` had no notion of a person, so 300+ visits could be 300 students or 30 students visiting ten times. Reports, contributions, and feedback arrived with no author, so admins could not follow up or spot repeat contributors. Favorites lived only in `localStorage`, so a student who switched from phone to laptop lost their starred courses.

Constraints that shaped the answer:

1. **Zero friction** — students open the site during a lecture to find a TD sheet. Anything that adds an email confirmation step or a password to forget loses most of them.
2. **Nothing sensitive behind the identity** — every link the site serves is already public to anyone who visits. There is no grade, no private file, no payment.
3. **Admin auth must not change** — Supabase Auth verifies admin credentials and the Go backend issues its own admin JWT (ADR 002). That flow works and protects real write access to content.
4. **Pre-signup activity should not be lost** — a visitor's first page view happens before they have any reason to tell us who they are.

Alternatives considered:

- **Stay anonymous, use a browser fingerprint or a random cookie ID** — no signup friction at all, but the identity dies with cleared storage, cannot be searched by name in the admin panel, and cannot sync favorites across devices
- **Email + password accounts** — real authentication, but adds password hashing, reset flows, email delivery, and abuse handling for a site whose content is public anyway; realistically it also cuts signups hard
- **Supabase Auth for students too** — reuses ADR 002 infrastructure, but puts student rows in a system we do not own the schema of, still requires email/password or a magic-link inbox round trip, and blurs the admin boundary we deliberately keep narrow
- **University email as the credential** — nice roster alignment, but requires deliverable mail per student and an unofficial site asking for a university address reads as phishing
- **Name + last name + a number 1-100** — a credential the student already knows, typed once, no inbox, no password; the number disambiguates the several `ali hassan`s in a cohort

## Decision

Introduce a **second, separate identity system for students** that identifies rather than authenticates. Admin auth stays exactly as ADR 002 describes.

**1. Identity is a triple.** A student is `first_name` + `last_name` + `number` (1-100). No email, no username, no password. The display handle is `first_last_number`, e.g. `mohamad_hassan_55`.

**2. New `users` table** (migration `000004_add_user_system`) with `id`, `first_name`, `last_name`, `number`, `is_guest`, `favorite_course_ids`, `created_at`, `last_seen_at`. Uniqueness is a **partial unique index** on `(first_name, last_name, number) WHERE is_guest = false`, plus a `CHECK` that `number` is between 1 and 100. Names are trimmed and lowercased on write so `Ali` and `ali` collide as they should.

**3. Guest-then-claim sessions.** A first-time visitor gets a guest row (`is_guest = true`) and a guest JWT before doing anything, so their page view is attributable. When they register, that **same row is claimed** — `UPDATE users SET ... WHERE id = $guest AND is_guest = true` — so the id, the `created_at`, and every pre-signup event stay attached to the person, and no duplicate visit is counted.

**4. App-issued student JWT**, signed with the same `JWT_SECRET`, carrying a user id claim and an is-guest flag, 365-day expiry, stored in `localStorage` under `infolinks_student_token`. Admin tokens use the same 365-day expiry, their own `admin: true` claim, and the `infolinks_token` key. A new `RequireUser` middleware sits alongside `RequireAdmin`; neither accepts the other's token.

**5. Activity carries `user_id`.** A nullable `user_id` FK is added to `page_views`, `link_clicks`, `reports`, `contributions`, and `feedback`. Nullable because rows written before this migration are genuinely anonymous and we do not want to invent an owner for them. A new append-only `favorite_events` table (`user_id`, `course_id`, `action` of `added` or `removed`, `created_at`) records favorites history, while `users.favorite_course_ids` stays the live set; both are written in one transaction.

**6. Gating.** Browsing stays fully open — search, expand a course, read the SEO pages, all anonymous as before. A **registered (non-guest)** session is required to open an external link, submit a report, contribution, or feedback, and to favorite a course. Guests can only record page views.

**7. HTTP contract.** Signup that hits the partial unique index returns `409` so the UI can tell the student to pick another number (55 becomes 65). Sign-in for a triple that does not exist returns `404` and points to signup.

**8. Analytics moves server-side.** The admin dashboard used to fetch entire `page_views` and `link_clicks` tables into the browser and count them in JavaScript. New endpoints aggregate in SQL instead — `GET /api/admin/analytics/summary` for unique-user counts (`COUNT(DISTINCT user_id)`) over today/7/30/90 days, `GET /api/admin/users` for a searchable student list, and `GET /api/admin/users/{id}` for a profile plus a unified activity timeline merging visits, link clicks, reports, contributions, feedback, and favorite events.

Student endpoints are `POST /api/users/guest`, `POST /api/users/register`, `POST /api/users/login`, `GET /api/users/me`, and `POST`/`DELETE /api/users/me/favorites/{course_id}`. Layering follows ADR 004 — handler, service, repository — and the schema change follows ADR 007.

## Consequences

- **This is identification, not authentication.** Anyone who knows a student's first name, last name, and number can sign in as them, see their favorites, and submit under their handle. There is no password to prove otherwise, and we are not pretending there is. It is an accepted trade-off: the site hosts public course links, stores nothing private behind the session, and the alternative — passwords — costs more signups than the risk is worth. If anything sensitive is ever put behind a student session, this decision has to be revisited first.
- Analytics finally distinguish people from page loads: unique registered students per range, top students by clicks, and per-student history.
- Reports, contributions, and feedback have a sender, so admins can recognize repeat contributors and follow up in Telegram.
- Favorites sync across devices for registered students instead of living in one browser.
- Pre-signup activity survives registration because the guest row is claimed rather than replaced. Sign-in from a guest browser **adopts** that guest: page views move onto the existing student and the empty guest row is deleted, so the first visit shows as the student's handle instead of `guest_<id>`.
- **Stale guests are deleted after 24 hours idle.** An hourly job (and one run at boot) removes `is_guest = true` rows whose `last_seen_at` is older than 24 hours. Their cascaded analytics go with them; registered students are never touched. A browser whose guest JWT points at a deleted row re-bootstraps a fresh guest on the next request.
- **Guests are exempt from the name constraint** because uniqueness is a partial index. Guest rows have null names and can pile up freely within the TTL window, which is what makes them cheap, but it also means the database alone does not stop two half-finished guests from looking identical.
- **Guests are counted as visits but not as people.** Unique-user metrics count registered students; guest traffic still shows up in raw visit counts and cannot be attributed to anyone.
- Two token types now share one `JWT_SECRET` and one middleware surface. The claims differ and each middleware checks its own, but it is one more thing to get right when touching auth.
- A 365-day token (student or admin) means a stolen or shared token stays valid for a year, and there is no revocation path short of rotating `JWT_SECRET`. Acceptable for students because this is identification, not a vault of secrets. For admins it is a real trade-off: leaked admin access lasts a year.
