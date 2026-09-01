// Normal load test: ramp up to 50 VUs, hold, ramp down
// Run: k6 run docs/load-test-normal.js
// X-Forwarded-For uses TEST-NET-3 (203.0.113.0/24), not 10.0.0.0/8 —
// private XFF hops are skipped by the rate limiter.
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 50 },
    { duration: "1m", target: 50 },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  const res = http.get("http://localhost:8080/api/content", {
    headers: { "X-Forwarded-For": `203.0.113.${__VU}` },
  });
  check(res, { "status is 200": (r) => r.status === 200 });
  sleep(1);
}
