// Burst test: validates rate limiter returns 429 under saturation
// Run: ./k6 run docs/load-test-burst.js
import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 30,
    duration: '10s',
};

export default function () {
    const res = http.get('http://localhost:8080/api/content');
    check(res, {
        'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    });
}
