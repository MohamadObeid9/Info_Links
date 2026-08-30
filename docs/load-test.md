# Load Test Results

## Environment

| | |
|---|---|
| **Date** | 2026-06-13 |
| **Go version** | 1.25.9 |
| **Stack** | Docker Compose (single container, app only — no separate DB container, connects to remote Supabase Postgres) |
| **Tool** | [k6](https://k6.io/) v0.56.0 |
| **Machine** | Fedora Linux (local dev) |
| **Endpoint tested** | `GET /api/content` |

---

## Test 1 — Normal Load (50 VUs, ramped)

**Script:** `docs/load-test-normal.js`  
**Scenario:** Ramp up to 50 virtual users over 30s, hold for 1m, ramp down over 10s. Each VU sends 1 req/s with a unique `X-Forwarded-For` IP to simulate 50 distinct users.

```
stages:
  30s → 50 VUs
  1m  → 50 VUs (hold)
  10s → 0 VUs
```

### Results

| Metric | Value |
|---|---|
| Total requests | 2,123 |
| Throughput | 21.1 req/s |
| Success rate | **100%** |
| Rate-limited (429) | 0 |
| avg latency | 902 ms |
| p(90) latency | 1.44 s |
| p(95) latency | 1.69 s |
| min latency | 316 ms |
| max latency | 3.77 s |
| Avg response size | ~36 KB |

**Threshold result:** `http_req_duration p(95) < 500ms` ❌ (p95 = 1.69s)  
**Failure rate threshold:** `rate < 1%` ✅ (0% failures)

### Analysis

The latency threshold failure is expected and understood. `/api/content` is a heavy read-all endpoint: a single CTE query aggregates programs, years, semesters, courses, links, extra sections, and extra links, returning ~36 KB of JSON per response. At 50 concurrent users all hitting this endpoint simultaneously, the remote Postgres connection pool queues up, and response times climb.

**This is not a real-world concern** because:
- The frontend caches the content response in `localStorage` for 1 hour. Real users hit `/api/content` once per hour at most, not once per second.
- At actual traffic levels (300 daily users), true concurrency on this endpoint is near zero.

**Fix shipped in production (2026-08-21):** origin sends `Cache-Control: public, max-age=60, stale-while-revalidate=600` on `GET /api/content`. Cloudflare caches that response (and hashed static assets). A cron ping every 10 minutes keeps the edge cache warm so students rarely wait on Postgres. Grafana p95/p99 for `/api/content` dropped from ~2–2.5s spikes to well under 500ms after that date.

---

## Test 2 — Burst (30 VUs, no sleep, 10s)

**Script:** `docs/load-test-burst.js`  
**Scenario:** 30 VUs hammering the same endpoint with no sleep for 10 seconds — all from the same IP. Validates that the rate limiter correctly returns 429 and does not crash or degrade.

### Results

| Metric | Value |
|---|---|
| Total requests | 53,070 |
| Throughput | **4,427 req/s** |
| Check pass rate | **100%** (all responses were 200 or 429) |
| Rate-limited (429) | 99.78% (52,956) |
| Passed through (200) | 0.22% (114) |
| p(95) latency (all) | 995 µs |
| p(95) latency (200 only) | 3.61 s |

### Analysis

The rate limiter works correctly. At 4,427 req/s from a single IP:
- The token bucket (10 tokens/s, burst 20) exhausts immediately and returns 429 in under 1ms without touching the DB or any business logic.
- The 114 requests that passed through represent the initial burst capacity (20 tokens) plus ~10 tokens/s refilled over the 10s window — exactly correct behaviour.
- Zero crashes, zero unexpected responses.

---

## Summary

| | Normal (50 VUs) | Burst (30 VUs, same IP) |
|---|---|---|
| Throughput | 21 req/s | 4,427 req/s |
| Success rate | 100% | 100% (200 + 429) |
| p(95) latency | 1.69 s | <1 ms (429) |
| Rate limiter effect | None (different IPs) | Correctly blocks 99.8% |

---

## Identified Bottleneck (June 2026 k6 run)

**`GET /api/content` under concurrent load** when every request hit origin Postgres.

Root cause: single heavy CTE query against remote Supabase, ~36 KB JSON, no edge cache at the time of the test.

**Production follow-up (August 2026):** Cloudflare in front of Render caches `/api/content` using the origin `Cache-Control` header. A 10-minute cron keeps the object warm. Grafana shows the step-change on 21 Aug: p95/p99 collapsed from multi-second spikes to a stable sub-500ms band. The k6 numbers above remain a valid origin-only baseline (k6 hits the Go process, not the CDN).

---

## Rate Limiter Trade-offs

The current implementation uses a `sync.Map` keyed by IP with one `*rate.Limiter` per IP (10 req/s, burst 20). Known limitations:

- **No eviction:** limiters are never removed from the map. Memory grows with the number of unique IPs seen since startup. At current traffic (300 daily users) this is negligible (<1 MB), but at scale a cleanup goroutine with a last-seen timestamp would be needed.
- **Single-process only:** state lives in-process. If horizontally scaled behind a load balancer, each instance has its own limiter map. A Redis-backed distributed rate limiter (e.g. `go-redis/redis_rate`) would be needed for correctness at scale.
- **Trusted proxy assumption:** `X-Forwarded-For` is trusted because all traffic arrives through Render's proxy. On a raw public server this header could be spoofed.
