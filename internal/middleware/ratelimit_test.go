package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:1234"

	// burn through the entire burst (20 tokens)
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
}

func TestRateLimit_DifferentIPsAreIsolated(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	store := newLimiterStore()

	store.get("1.1.1.1")
	store.get("2.2.2.2")
	if store.len() != 2 {
		t.Fatalf("expected 2 entries, got %d", store.len())
	}

	// Make one entry look idle by backdating its lastSeen.
	store.mu.Lock()
	store.limiters["1.1.1.1"].lastSeen = time.Now().Add(-2 * limiterIdleTTL)
	store.mu.Unlock()

	removed := store.evictIdle(time.Now().Add(-limiterIdleTTL))
	if removed != 1 {
		t.Fatalf("expected 1 entry evicted, got %d", removed)
	}
	if store.len() != 1 {
		t.Fatalf("expected 1 entry remaining, got %d", store.len())
	}
	if _, ok := store.limiters["2.2.2.2"]; !ok {
		t.Error("active entry should not have been evicted")
	}
}

func TestLimiterStore_GetReusesLimiter(t *testing.T) {
	store := newLimiterStore()

	first := store.get("9.9.9.9")
	second := store.get("9.9.9.9")
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
