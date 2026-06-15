package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
}

const (
	rateLimitRPS      = 10
	rateLimitBurst    = 20
	limiterIdleTTL    = 10 * time.Minute
	limiterSweepEvery = 5 * time.Minute
)

var trustedProxyNets []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8", "::1/128",
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"169.254.0.0/16", "fc00::/7", "fe80::/10",
	} {
		if _, n, err := net.ParseCIDR(cidr); err == nil {
			trustedProxyNets = append(trustedProxyNets, n)
		}
	}
}

func RateLimit(next http.Handler) http.Handler {
	store := newLimiterStore()

	go func() {
		ticker := time.NewTicker(limiterSweepEvery)
		defer ticker.Stop()
		for range ticker.C {
			store.evictIdle(time.Now().Add(-limiterIdleTTL))
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isExempt(r) || store.get(getClientIP(r)).Allow() {
			next.ServeHTTP(w, r)
			return
		}
		writeJSONErr(w, http.StatusTooManyRequests, "rate limit exceeded")
	})
}

func isExempt(r *http.Request) bool {
	p := r.URL.Path
	switch p {
	case "/healthz", "/readyz", "/metrics", "/robots.txt":
		return true
	}
	if strings.HasPrefix(p, "/assets/") {
		return true
	}
	dot := strings.LastIndexByte(p, '.')
	return dot != -1 && !strings.ContainsRune(p[dot:], '/')
}

func getClientIP(r *http.Request) string {
	remote := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remote); err == nil {
		remote = host
	}
	if isTrustedProxy(net.ParseIP(remote)) {
		if ip := clientFromXFF(r.Header.Get("X-Forwarded-For")); ip != "" {
			return ip
		}
	}
	return remote
}

func clientFromXFF(header string) string {
	parts := strings.Split(header, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		if ip := net.ParseIP(candidate); ip != nil && !isTrustedProxy(ip) {
			return candidate
		}
	}
	return ""
}

func isTrustedProxy(ip net.IP) bool {
	for _, n := range trustedProxyNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func newLimiterStore() *limiterStore {
	return &limiterStore{limiters: make(map[string]*limiterEntry)}
}

func (s *limiterStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.limiters[key]
	if !ok {
		e = &limiterEntry{limiter: rate.NewLimiter(rate.Limit(rateLimitRPS), rateLimitBurst)}
		s.limiters[key] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func (s *limiterStore) evictIdle(cutoff time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for key, e := range s.limiters {
		if e.lastSeen.Before(cutoff) {
			delete(s.limiters, key)
			removed++
		}
	}
	return removed
}

func (s *limiterStore) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.limiters)
}
