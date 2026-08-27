## Users (Student Identity) in This Project

### What this slice is

Students identify themselves with **first name + last name + a number 1-100**. No email, no username, no password. The handle they see is `first_last_number`, e.g. `mohamad_hassan_55`.

This is a **second identity system**, deliberately separate from admin auth. Admin login still goes through Supabase and gets an `admin: true` JWT (ADR 002). Student sessions are app-issued JWTs carrying a user id. They share `JWT_SECRET` and nothing else.

The reason it exists: without a person attached to a row, `page_views` and `link_clicks` can tell you how many hits there were but not how many humans. See [ADR 008](../adr/008-student-identity-without-passwords.md) for the decision and its trade-offs — the honest one being that this identifies students, it does not authenticate them.

---

### The pieces

| Layer | File | Owns |
|---|---|---|
| Handler | `internal/api/users_handlers.go` | Decode body, read claims, issue JWT, map errors to status codes |
| Service | `internal/service/users.go` | Normalize names, validate number range, choose claim vs create |
| Repository | `internal/repository/users.go` | SQL, `23505` detection, zero-row detection |
| SQL | `internal/repository/queries.go` | Query constants |
| Errors | `internal/errs/errors.go` | `ErrUsernameTaken`, `ErrUserGuestNotFound`, `ErrUserNumberRange`, `ErrUserNameRequired` |
| Middleware | `internal/middleware/auth_middleware.go` | `RequireUser` beside `RequireAdmin` |

Same layering as every other domain (ADR 004). Nothing here is special except the claim path.

---

### The flow, end to end

A first-time visitor has no session, so the frontend bootstraps one before doing anything else:

```text
POST /api/users/guest
  → UserService.CreateGuest
  → INSERT INTO users (is_guest) VALUES (true) RETURNING id
  → generateUserToken(id, isGuest = true)
  → 201 { token }
```

Now their page view has an owner:

```text
POST /api/page_views  + Authorization: Bearer <student token>
  → user_id taken from the token claims, never from the body
```

Later they try to open a link, which is gated. The frontend shows the signup modal and calls register with the **guest token still attached**:

```text
POST /api/users/register + Bearer <guest token>
  → handler reads user_id and is_guest from claims
  → UserService.RegisterUser(ctx, guestID, user)
  → is_guest in claims == true → repo.ClaimGuest
  → UPDATE users SET first_name, last_name, number, is_guest = false, last_seen_at = now()
    WHERE id = $guest AND is_guest = true
    RETURNING id
  → new token for the SAME id
```

The row is updated, not replaced. Same `id`, same `created_at`, and the page views recorded minutes earlier now belong to a named student. No second visit is counted for the same person.

A returning student signs in instead — `POST /api/users/login` looks up the triple, returns a token on hit, `404` on miss. If this browser still has a guest token (first visit already recorded), the handler passes that guest id in and the service **adopts** it: every `page_views` / activity row moves from the guest onto the student, then the empty guest is deleted. Admin "Who visited today" then shows `mohamad_hassan_55` instead of `guest_42`. The browser does not send a second page view (`sessionStorage` already marked this tab), so without adopt the visit would stay a ghost guest forever.

The once-per-tab visit guard (`sessionStorage.pv_tracked`) is keyed to the **user id**, not a bare boolean. If the guest token is rejected and a fresh guest is bootstrapped, the old guard is cleared and a new page view is recorded for the new id. Otherwise the visit stays on a dead `guest_N` row while the student who later signs up looks active (link clicks, last seen) but never appears in "Who visited today". That list also counts people with a link click today, not only `page_views`, so a click without a visit row still surfaces the person.

---

### Why the claim decision must come from the JWT

This is the one rule in this slice worth burning into memory.

`RegisterUser` decides between **claim an existing guest row** and **create a brand new row**. If that decision reads an `is_guest` flag out of the request body, the endpoint becomes:

```json
{ "first_name": "ali", "last_name": "hassan", "number": 55, "is_guest": true, "guest_id": 91 }
```

...and anyone can post that. Claiming is an `UPDATE` on a row identified by id, so a body-supplied guest id lets a caller overwrite someone else's row with their own name — and if that row was already a registered student, you have handed out an account takeover. The `AND is_guest = true` in the WHERE clause blunts it (a claimed row can never be re-claimed), but the guest id itself is still attacker-chosen.

The rule: **the guest id and the is-guest flag come from the verified token claims**. The body carries only what the student typed — first name, last name, number. A client cannot forge claims without `JWT_SECRET`; it can put anything it likes in a JSON body.

Practically, that means the handler does the claim extraction and the service takes `guestID` as an argument it can trust:

```go
func (s *UserService) RegisterUser(ctx context.Context, guestID int, u models.User) (models.User, error)
```

The service does not read `u.IsGuest`. Any `is_guest` field arriving in the body is ignored (`decodeJSONBody` uses `DisallowUnknownFields`, so unexpected keys are rejected outright).

Same principle everywhere else in this slice: `user_id` on a page view, link click, report, contribution, or feedback is always the id from the token, never a body field. Otherwise anyone can write history into anyone's timeline and the analytics are fiction.

---

### Why names are normalized on write

The database enforces uniqueness with a partial index:

```sql
CREATE UNIQUE INDEX users_unique_username
    ON public.users (first_name, last_name, number)
    WHERE is_guest = false;
```

Postgres compares text exactly. `Ali` and `ali` and ` ali ` are three different keys, so without normalization the index happily accepts all three and the collision check silently stops working — students end up with near-duplicate identities and sign-in fails for the one who typed their name with a different capitalization than at signup.

So the service trims and lowercases before anything touches SQL:

```go
u.FirstName = strings.ToLower(strings.TrimSpace(u.FirstName))
u.LastName  = strings.ToLower(strings.TrimSpace(u.LastName))
```

Both register and login run the same normalization, which is the part that matters — a lookup normalized differently from the write is a lookup that never finds the row. Normalizing in the service (not the repository, not the handler) keeps it in one place that both paths already go through, and matches how the rest of the project treats input cleaning.

The index is `WHERE is_guest = false` because guest rows have null names and there may be thousands of them; they must not compete for uniqueness. The trade-off: the database constrains registered students only.

---

### Why `23505` becomes a sentinel error

Two students really can pick `ali hassan 55`. The second one must get a clean `409` with "pick another number", not a 500.

Checking "does this triple exist?" before inserting does not work — two requests can both pass the check and then both insert. The unique index is the only thing that actually decides, and it decides at write time. So the repository lets the write fail and translates the failure:

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    return 0, errs.ErrUsernameTaken
}
return 0, fmt.Errorf("create user: %w", err)
```

`23505` is Postgres's `unique_violation`. Turning it into `errs.ErrUsernameTaken` at the repository boundary is the same contract the rest of the project uses — the repository converts database outcomes into domain outcomes, the service passes them up, the handler maps them to HTTP:

| Outcome | Where detected | Sentinel | Status |
|---|---|---|---|
| Triple already registered | `23505` from the unique index | `ErrUsernameTaken` | `409` |
| Claim hit no guest row | `sql.ErrNoRows` on the `UPDATE ... RETURNING` | `ErrUserGuestNotFound` | `404` (or re-bootstrap) |
| Sign-in triple unknown | `sql.ErrNoRows` on the lookup | `ErrUserNotFound` | `404` |
| Empty name after trimming | Service validation | `ErrUserNameRequired` | `400` |
| Number outside 1-100 | Service validation | `ErrUserNumberRange` | `400` |
| Anything else | Repository | wrapped with `%w` | log + `500` |

Note the claim path detects **zero rows via `sql.ErrNoRows`**, because it uses `UPDATE ... RETURNING id` and scans the result. A claim that matches nothing means the id is bogus or the row was already claimed — either way the client's token is stale and it should bootstrap a fresh guest.

The number range is checked twice on purpose: in Go so the student gets a readable 400, and as a `CHECK` constraint in the schema so no code path can ever write a 0 or a 4000. Validation in the application is a message; validation in the database is a guarantee.

---

### Why favorites need a transaction

Favorites are stored twice, on purpose:

- `users.favorite_course_ids` — the **live set**, so "My Courses" is one row read
- `favorite_events` — an **append-only log** of `added` / `removed` with timestamps, so the admin timeline can say "on 12 May added Algorithms, removed Databases"

An array alone loses history. A log alone means replaying every event to render a star icon. Keeping both is a deliberate duplication, and duplication is only safe if the two copies can never disagree.

So a toggle is one transaction:

```go
tx, err := r.db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("add favorite: %w", err)
}
defer tx.Rollback()

// 1. array_append (or array_remove) on users.favorite_course_ids
// 2. INSERT INTO favorite_events (user_id, course_id, action)

return tx.Commit()
```

Without it, a crash between the two writes leaves either a starred course with no event (the timeline lies about when it happened) or an event with no star (the student's list silently drops a course). Neither is recoverable afterwards, because there is no third source to reconcile against.

`defer tx.Rollback()` after a successful `Commit()` is a no-op returning `sql.ErrTxDone`, which is why it is safe to defer unconditionally.

The array update is also written to be idempotent — favoriting an already-favorited course must not append a duplicate id. Guard on the array side (`array_append` only when the id is absent) rather than trusting the client not to double-fire.

---

### Gating: who may do what

| Action | Anonymous | Guest | Registered |
|---|---|---|---|
| Browse, search, expand courses | yes | yes | yes |
| Page view recorded | not attributed | yes | yes |
| Open an external link | no | no | yes |
| Report / contribute / feedback | no | no | yes |
| Favorite a course | no | no | yes |

`RequireUser` validates the token and rejects a missing or invalid one. The endpoints that need a real student additionally reject `is_guest = true` — a valid token is not the same as a known person, and forgetting that check is the easy way to let guests write reports nobody can trace.

Admin requests keep skipping analytics inserts, so admin browsing does not pollute student metrics.

---

### Common mistakes to avoid

- Trusting `user_id`, `guest_id`, or `is_guest` from a request body instead of the token claims
- Normalizing names in register but not in login (the row exists and sign-in still 404s)
- Doing a `SELECT` existence check instead of letting `23505` decide, then wondering about the occasional duplicate under concurrency
- Writing `favorite_course_ids` and `favorite_events` outside one transaction
- Treating `COUNT(*) FROM users` as the student count — filter `WHERE is_guest = false`
- Accepting a guest token on a route that must attribute activity to a named person
- Assuming unclaimed guests live forever — idle guests (`last_seen_at` older than 24h) are deleted by the server cleanup job; their analytics cascade away
