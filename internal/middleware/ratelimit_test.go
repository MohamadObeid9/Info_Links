package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "2.3.4.5:9999"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimit_BlocksWhenExceeded(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:1234"

	// burn through the entire default burst (20 tokens)
	for i := 0; i < rateLimitBurst; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// this one should be blocked
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimit_DifferentIPsAreIsolated(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// burn IP A
	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "203.0.113.1:1111"
	for i := 0; i < rateLimitBurst; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, reqA)
	}

	// IP B should still be allowed
	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "203.0.113.2:2222"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqB)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for different IP, got %d", rr.Code)
	}
}

// TestRateLimit_SpoofedXFFFromUntrustedPeerCannotBypass verifies that a client
// connecting directly (untrusted RemoteAddr) cannot evade the limiter by
// sending a unique X-Forwarded-For per request: all requests must key off the
// same RemoteAddr.
func TestRateLimit_SpoofedXFFFromUntrustedPeerCannotBypass(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const remote = "203.0.113.9:5555"

	// Burn the burst while rotating a spoofed XFF on every request.
	for i := 0; i < rateLimitBurst; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		req.Header.Set("X-Forwarded-For", randomIPv4(i))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// The next request, still from the same untrusted peer with yet another
	// spoofed XFF, must be blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remote
	req.Header.Set("X-Forwarded-For", randomIPv4(999))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("spoofed XFF bypassed the limiter: expected 429, got %d", rr.Code)
	}
}

func TestRateLimit_AdminAuthStricterThanDefault(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "198.51.100.10:4444"

	for i := 0; i < adminAuthBurst; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d within burst: expected 200, got %d", i+1, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after admin auth burst, got %d", rr.Code)
	}
}

func TestRateLimit_WriteUserStricterThanDefault(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/contributions", nil)
	req.RemoteAddr = "198.51.100.11:4444"

	for i := 0; i < writeUserBurst; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d within burst: expected 200, got %d", i+1, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after write-user burst, got %d", rr.Code)
	}
}

func TestRateLimit_ClassesAreIsolated(t *testing.T) {
	handler := RateLimit("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const remote = "198.51.100.12:4444"

	authReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	authReq.RemoteAddr = remote
	for i := 0; i < adminAuthBurst; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, authReq)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, authReq)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected auth class exhausted, got %d", rr.Code)
	}

	// Default browsing from the same IP should still work.
	pageReq := httptest.NewRequest(http.MethodGet, "/api/content", nil)
	pageReq.RemoteAddr = remote
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, pageReq)
	if rr.Code != http.StatusOK {
		t.Errorf("default class should remain available, got %d", rr.Code)
	}
}

func TestRateLimit_CooldownAfterRepeatedDenials(t *testing.T) {
	now := time.Now()
	cfg := rateLimitConfig{
		limits: map[routeClass]classLimit{
			classDefault:   {rate.Limit(100), 1},
			classAdminAuth: {rate.Limit(1), 1},
			classAdminAPI:  {rate.Limit(1), 1},
			classIdentity:  {rate.Limit(1), 1},
			classWriteUser: {rate.Limit(1), 1},
		},
		cooldownN:   3,
		cooldownFor: 2 * time.Minute,
		idleTTL:     limiterIdleTTL,
		sweepEvery:  time.Hour,
		now:         func() time.Time { return now },
	}

	handler := newRateLimitHandler("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "198.51.100.20:9999"

	// 1 allowed (burst), then denials until cooldown trips.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", rr.Code)
	}

	for i := 0; i < 3; i++ {
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("denial %d: expected 429, got %d", i+1, rr.Code)
		}
	}

	// Cooldown engaged: still 429 with Retry-After reflecting cooldown.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected cooldown 429, got %d", rr.Code)
	}
	retryAfter, err := strconv.Atoi(rr.Header().Get("Retry-After"))
	if err != nil {
		t.Fatalf("Retry-After: %v", err)
	}
	if retryAfter < 60 {
		t.Errorf("expected Retry-After near cooldown duration, got %d", retryAfter)
	}

	// After cooldown expires, requests are allowed again.
	now = now.Add(3 * time.Minute)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("after cooldown expected 200, got %d", rr.Code)
	}
}

func TestRateLimit_AuthenticatedAdminUsesDefaultClass(t *testing.T) {
	const secret = "rate-limit-admin-test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"admin": true})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := RateLimit(secret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.RemoteAddr = "198.51.100.30:4444"
	req.Header.Set("Authorization", "Bearer "+signed)

	// More than admin_api burst (5), still under default burst (20).
	for i := 0; i < adminAPIBurst+3; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("authed admin request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRouteClassFor_AdminProbeVsAuthed(t *testing.T) {
	const secret = "rate-limit-admin-class-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"admin": true})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	probe := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	if got := routeClassFor(probe, secret); got != classAdminAPI {
		t.Errorf("probe: got %q want %q", got, classAdminAPI)
	}

	authed := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	authed.Header.Set("Authorization", "Bearer "+signed)
	if got := routeClassFor(authed, secret); got != classDefault {
		t.Errorf("authed admin: got %q want %q", got, classDefault)
	}
}

func TestRouteClassFor(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   routeClass
	}{
		{http.MethodPost, "/api/auth/login", classAdminAuth},
		{http.MethodGet, "/api/admin/users", classAdminAPI},
		{http.MethodPost, "/api/admin/links", classAdminAPI},
		{http.MethodPost, "/api/users/guest", classIdentity},
		{http.MethodPost, "/api/users/register", classIdentity},
		{http.MethodPost, "/api/users/login", classIdentity},
		{http.MethodPost, "/api/contributions", classWriteUser},
		{http.MethodPost, "/api/reports", classWriteUser},
		{http.MethodPost, "/api/feedback", classWriteUser},
		{http.MethodGet, "/api/content", classDefault},
		{http.MethodPost, "/api/page_views", classDefault},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if got := routeClassFor(req, ""); got != tt.want {
				t.Errorf("routeClassFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{
			name:       "no XFF uses RemoteAddr",
			remoteAddr: "203.0.113.5:443",
			want:       "203.0.113.5",
		},
		{
			name:       "untrusted peer ignores XFF",
			remoteAddr: "203.0.113.5:443",
			xff:        "1.2.3.4",
			want:       "203.0.113.5",
		},
		{
			name:       "trusted peer honors single XFF entry",
			remoteAddr: "10.0.0.7:5555",
			xff:        "198.51.100.23",
			want:       "198.51.100.23",
		},
		{
			name:       "trusted peer uses rightmost non-proxy entry",
			remoteAddr: "10.0.0.7:5555",
			xff:        "198.51.100.23, 10.0.0.8",
			want:       "198.51.100.23",
		},
		{
			name:       "trusted peer skips spoofed leftmost client values",
			remoteAddr: "127.0.0.1:80",
			xff:        "spoofed, 203.0.113.77, 10.0.0.9",
			want:       "203.0.113.77",
		},
		{
			name:       "trusted peer with only proxy entries falls back to RemoteAddr",
			remoteAddr: "10.0.0.7:5555",
			xff:        "10.0.0.8, 192.168.1.1",
			want:       "10.0.0.7",
		},
		{
			name:       "trusted peer with empty XFF falls back to RemoteAddr",
			remoteAddr: "10.0.0.7:5555",
			xff:        "   ",
			want:       "10.0.0.7",
		},
		{
			name:       "trusted peer with garbage XFF falls back to RemoteAddr",
			remoteAddr: "10.0.0.7:5555",
			xff:        "not-an-ip",
			want:       "10.0.0.7",
		},
		{
			name:       "RemoteAddr without port returned as-is",
			remoteAddr: "203.0.113.5",
			want:       "203.0.113.5",
		},
		{
			name:       "IPv6 trusted loopback honors XFF",
			remoteAddr: "[::1]:8080",
			xff:        "198.51.100.42",
			want:       "198.51.100.42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			if got := getClientIP(req); got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientFromXFF(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "empty", header: "", want: ""},
		{name: "single public", header: "198.51.100.1", want: "198.51.100.1"},
		{name: "rightmost real client", header: "198.51.100.1, 10.0.0.2", want: "198.51.100.1"},
		{name: "all proxies", header: "10.0.0.1, 192.168.0.1", want: ""},
		{name: "invalid entries skipped", header: "garbage, , 198.51.100.9", want: "198.51.100.9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientFromXFF(tt.header); got != tt.want {
				t.Errorf("clientFromXFF(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestLimiterStore_EvictsIdleEntries(t *testing.T) {
	store := newLimiterStore(productionRateLimitConfig())

	store.get("1.1.1.1:"+string(classDefault), classDefault)
	store.get("2.2.2.2:"+string(classDefault), classDefault)
	if store.len() != 2 {
		t.Fatalf("expected 2 entries, got %d", store.len())
	}

	// Make one entry look idle by backdating its lastSeen.
	store.mu.Lock()
	store.limiters["1.1.1.1:"+string(classDefault)].lastSeen = time.Now().Add(-2 * limiterIdleTTL)
	store.mu.Unlock()

	removed := store.evictIdle(time.Now().Add(-limiterIdleTTL))
	if removed != 1 {
		t.Fatalf("expected 1 entry evicted, got %d", removed)
	}
	if store.len() != 1 {
		t.Fatalf("expected 1 entry remaining, got %d", store.len())
	}
	if _, ok := store.limiters["2.2.2.2:"+string(classDefault)]; !ok {
		t.Error("active entry should not have been evicted")
	}
}

func TestLimiterStore_GetReusesLimiter(t *testing.T) {
	store := newLimiterStore(productionRateLimitConfig())

	first := store.get("9.9.9.9:"+string(classDefault), classDefault)
	second := store.get("9.9.9.9:"+string(classDefault), classDefault)
	if first != second {
		t.Error("expected the same *rate.Limiter to be reused for the same key")
	}
	if store.len() != 1 {
		t.Fatalf("expected 1 entry, got %d", store.len())
	}
}

// randomIPv4 builds a distinct public IPv4 string per seed so each request can
// carry a unique spoofed X-Forwarded-For value.
func randomIPv4(seed int) string {
	return fmt.Sprintf("11.%d.%d.%d", seed%256, (seed*7)%256, (seed*13+1)%256)
}
