# Load Test Results

## Environment (2026-09-01, after in-memory cache)

| | |
|---|---|
| **Date** | 2026-09-01 (afternoon rerun) |
| **Go version** | 1.26.5 |
| **Stack** | Local `go run ./cmd/server` with `APP_ENV=production` (remote Supabase). Origin now keeps a 60s in-memory copy of `GET /api/content` with `singleflight` on miss. |
| **Tool** | [k6](https://k6.io/) v2.2.0 |
| **Machine** | Fedora Linux 44 (local dev) |
| **Endpoint tested** | `GET /api/content` |

Origin-only (k6 hits the Go process, not Cloudflare).

---

## Test 1 — Normal Load (50 VUs, ramped)

**Script:** `docs/load-test-normal.js`  
**Scenario:** Ramp up to 50 virtual users over 30s, hold for 1m, ramp down over 10s. Each VU sends ~1 req/s with a unique `X-Forwarded-For` (`203.0.113.{VU}`, TEST-NET-3).

```
stages:
  30s → 50 VUs
  1m  → 50 VUs (hold)
  10s → 0 VUs
```

### Results (2026-09-01 afternoon — in-memory cache)

| Metric | Value |
|---|---|
| Total requests | 4,018 |
| Throughput | 39.9 req/s |
| Success rate (HTTP 200) | **100%** |
| Rate-limited (429) | 0 |
| avg latency | 1.76 ms |
| p(90) latency | 1.23 ms |
| p(95) latency | **1.53 ms** |
| min latency | 317 µs |
| max latency | 329 ms |

**Threshold result:** `http_req_duration p(95) < 500ms` ✅  
**Failure rate threshold:** `rate < 1%` ✅ (0%)

Same day, before the cache (morning): 1,210 req, 11.9 req/s, p95 **4.91 s**, 99.5% 200. June 2026 (no cache): 2,123 req, 21.1 req/s, p95 1.69 s.

After warmup, 50 distinct IPs no longer stampede Postgres. Max 329 ms is a refill after TTL/cold; the rest is RAM.

---

## Test 2 — Burst (30 VUs, no sleep, 10s)

**Script:** `docs/load-test-burst.js`  
**Scenario:** 30 VUs, same IP, 10 seconds — rate limiter still returns 429.

### Results (2026-09-01 afternoon)

| Metric | Value |
|---|---|
| Total requests | 170,119 |
| Throughput | **17,009 req/s** |
| Check pass rate | **100%** (200 or 429) |
| Rate-limited (429) | 99.93% (170,000) |
| Passed through (200) | 0.07% (119) |
| p(95) latency (all) | 3.8 ms |
| p(95) latency (200 only) | **8.91 ms** |

Morning (no cache): 167,584 req, 15,505 req/s, 119 × 200, p95 of 200s **2.91 s**. The 200-count is still the token bucket (burst 20 + ~10/s). Those 200s are now served from RAM, so p95 for allowed requests dropped from seconds to milliseconds. 429 behaviour is unchanged.

---

## Summary

| | Normal (cache) | Normal (same day, no cache) | Burst (cache) | Burst (same day, no cache) |
|---|---|---|---|---|
| Throughput | 40 req/s | 12 req/s | 17,009 req/s | 15,505 req/s |
| Success | 100% 200 | 99.5% 200 | 100% (200 + 429) | 100% (200 + 429) |
| p(95) | **1.53 ms** | 4.91 s | 3.8 ms (all) | 2.1 ms (all) |
| p(95) of 200s | 1.53 ms | 4.91 s | **8.91 ms** | 2.91 s |
| Rate limiter | none | 6 × 429 | 99.93% 429 | 99.92% 429 |

---

## Bottleneck

The origin CTE against Supabase is still expensive on a **cold miss**. Under synthetic concurrency it no longer runs once per request: the in-memory cache + `singleflight` collapses that to one query per TTL (or after admin `Invalidate()`).

Cloudflare in production still absorbs student traffic at the edge. This k6 path is origin-only and now matches “warm RAM” rather than “50 concurrent Postgres round-trips.”

---

## Rate limiter

Unchanged: 10 req/s, burst 20, per IP; idle eviction after 10 minutes. Burst test still proves 429s do not crash the process.
