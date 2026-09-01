# ADR 010: In-memory cache for `GET /api/content`

## Status

Accepted (2026-09-01)

## Context

`GET /api/content` builds the full course tree with a heavy Postgres JSON aggregation. Cloudflare already caches the response at the edge for students. Origin-only traffic (k6, a Cloudflare miss, a flood of distinct IPs) still ran that CTE **once per request**.

A k6 normal test (50 VUs, unique `X-Forwarded-For`) on 2026-09-01 morning hit remote Supabase directly: p95 **4.91 s**, 99.5% HTTP 200.

## Decision

Keep a **process-local** copy of the public content payload:

- 60s TTL
- `singleflight` so concurrent misses share one query
- public `GET /api/content` reads the cache (`Get`)
- admin `GET /api/admin/content` bypasses it (`GetUncached`) and still freezes expired services
- successful mutations of courses, links, extra sections/links, and services call `Invalidate()`

Cloudflare + the 10-minute warm-up cron stay in place. This cache is for the origin, not a replacement for the CDN.

## Consequences

- Origin k6 (2026-09-01 afternoon): normal load **100% 200**, p95 **1.53 ms**; burst **~17,009 req/s**, 99.93% 429, p95 of allowed 200s **8.91 ms**. Full tables in [`docs/load-test.md`](../load-test.md).
- A cold miss or TTL expiry still pays for the CTE once (max ~329 ms in that run).
- Multi-instance deploys can serve stale JSON for up to 60s unless every instance is invalidated; one Render web service is enough at current scale.
- Admin UI is not served from this cache.
