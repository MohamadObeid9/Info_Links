import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 50 }, // ramp up to 50 VUs
    { duration: "1m", target: 50 }, // hold at 50 VUs
    { duration: "10s", target: 0 }, // ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests under 500ms
    http_req_failed: ["rate<0.01"], // less than 1% errors
  },
};

export default function () {
  const res = http.get("http://localhost:8080/api/content");
  check(res, { "status is 200": (r) => r.status === 200 });
  sleep(1);
}
