package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type routeClass string

const (
	classDefault   routeClass = "default"
	classAdminAuth routeClass = "admin_auth"
	classAdminAPI  routeClass = "admin_api"
	classIdentity  routeClass = "identity"
	classWriteUser routeClass = "write_user"
)

type classLimit struct {
	rps   rate.Limit
	burst int
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type cooldownEntry struct {
	denials       int
	cooldownUntil time.Time
	lastSeen      time.Time
}

type limiterStore struct {
	mu          sync.Mutex
	limiters    map[string]*limiterEntry
	cooldowns   map[string]*cooldownEntry
	limits      map[routeClass]classLimit
	cooldownN   int
	cooldownFor time.Duration
	now         func() time.Time
}

type rateLimitConfig struct {
	limits      map[routeClass]classLimit
	cooldownN   int
	cooldownFor time.Duration
	idleTTL     time.Duration
	sweepEvery  time.Duration
	now         func() time.Time
}

const (
	rateLimitRPS      = 10
	rateLimitBurst    = 20
	limiterIdleTTL    = 10 * time.Minute
	limiterSweepEvery = 5 * time.Minute

	adminAuthRPS   = 1
	adminAuthBurst = 3
	adminAPIRPS    = 2
	adminAPIBurst  = 5
	identityRPS    = 2
	identityBurst  = 5
	writeUserRPS   = 1
	writeUserBurst = 3

	cooldownAfterDenials = 10
	cooldownDuration     = 15 * time.Minute
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

func productionRateLimitConfig() rateLimitConfig {
	return rateLimitConfig{
		limits: map[routeClass]classLimit{
			classDefault:   {rate.Limit(rateLimitRPS), rateLimitBurst},
			classAdminAuth: {rate.Limit(adminAuthRPS), adminAuthBurst},
			classAdminAPI:  {rate.Limit(adminAPIRPS), adminAPIBurst},
			classIdentity:  {rate.Limit(identityRPS), identityBurst},
			classWriteUser: {rate.Limit(writeUserRPS), writeUserBurst},
		},
		cooldownN:   cooldownAfterDenials,
		cooldownFor: cooldownDuration,
		idleTTL:     limiterIdleTTL,
		sweepEvery:  limiterSweepEvery,
		now:         time.Now,
	}
}

func RateLimit(jwtSecret string, next http.Handler) http.Handler {
	return newRateLimitHandler(jwtSecret, next, productionRateLimitConfig())
}

func newRateLimitHandler(jwtSecret string, next http.Handler, cfg rateLimitConfig) http.Handler {
	if cfg.now == nil {
		cfg.now = time.Now
	}
	store := newLimiterStore(cfg)

	go func() {
		ticker := time.NewTicker(cfg.sweepEvery)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := cfg.now().Add(-cfg.idleTTL)
			store.evictIdle(cutoff)
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isExempt(r) {
			next.ServeHTTP(w, r)
			return
		}

		ip := getClientIP(r)
		class := routeClassFor(r, jwtSecret)
		key := ip + ":" + string(class)

		if retryAfter, blocked := store.cooldownActive(key); blocked {
			writeRateLimited(w, retryAfter)
			return
		}

		if store.get(key, class).Allow() {
			next.ServeHTTP(w, r)
			return
		}

		retryAfter := 1
		if class != classDefault {
			retryAfter = store.recordDenial(key)
		}
		writeRateLimited(w, retryAfter)
	})
}

func writeRateLimited(w http.ResponseWriter, retryAfterSec int) {
	if retryAfterSec < 1 {
		retryAfterSec = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
	writeJSONErr(w, http.StatusTooManyRequests, "rate limit exceeded")
}

func routeClassFor(r *http.Request, jwtSecret string) routeClass {
	p := r.URL.Path
	if r.Method == http.MethodPost && p == "/api/auth/login" {
		return classAdminAuth
	}
	if strings.HasPrefix(p, "/api/admin/") {
		// Real admin sessions load many endpoints at once; keep the strict
		// bucket only for unauthenticated / non-admin probes.
		if jwtSecret != "" && IsAuthenticatedAdmin(jwtSecret, r.Header.Get("Authorization")) {
			return classDefault
		}
		return classAdminAPI
	}
	if r.Method == http.MethodPost {
		switch p {
		case "/api/users/guest", "/api/users/register", "/api/users/login":
			return classIdentity
		case "/api/contributions", "/api/reports", "/api/feedback":
			return classWriteUser
		}
	}
	return classDefault
}

func isExempt(r *http.Request) bool {
	p := r.URL.Path
	switch p {
	case "/healthz", "/readyz", "/metrics", "/robots.txt", "sitemap.xml", "/.well-known/api-catalog", "/.well-known/oauth-protected-resource", "/.well-known/oauth-authorization-server", "/.well-known/openid-configuration", "/.well-known/jwks.json", "/.well-known/agent-card.json", "/.well-known/agents-index.json", "/.well-known/agent-skills/index.json", "/.well-known/mcp/server-card.json", "/.well-known/http-message-signatures-directory", "/openapi.json", "/auth.md", "/mcp":
		return true
	}
	if strings.HasPrefix(p, "/.well-known/agent-skills/") {
		return true
	}
	if strings.HasPrefix(p, "/.well-known/mcp/") {
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

func newLimiterStore(cfg rateLimitConfig) *limiterStore {
	return &limiterStore{
		limiters:    make(map[string]*limiterEntry),
		cooldowns:   make(map[string]*cooldownEntry),
		limits:      cfg.limits,
		cooldownN:   cfg.cooldownN,
		cooldownFor: cfg.cooldownFor,
		now:         cfg.now,
	}
}

func (s *limiterStore) get(key string, class routeClass) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.limiters[key]
	if !ok {
		lim := s.limits[class]
		if lim.burst == 0 {
			lim = s.limits[classDefault]
		}
		e = &limiterEntry{limiter: rate.NewLimiter(lim.rps, lim.burst)}
		s.limiters[key] = e
	}
	e.lastSeen = s.now()
	return e.limiter
}

func (s *limiterStore) cooldownActive(key string) (retryAfterSec int, blocked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cooldowns[key]
	if !ok || c.cooldownUntil.IsZero() {
		return 0, false
	}
	now := s.now()
	c.lastSeen = now
	if now.Before(c.cooldownUntil) {
		return int(c.cooldownUntil.Sub(now).Seconds()) + 1, true
	}
	// Cooldown expired: clear abuse state and give the IP a fresh token bucket.
	delete(s.cooldowns, key)
	delete(s.limiters, key)
	return 0, false
}

func (s *limiterStore) recordDenial(key string) (retryAfterSec int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	c, ok := s.cooldowns[key]
	if !ok {
		c = &cooldownEntry{}
		s.cooldowns[key] = c
	}
	c.lastSeen = now
	c.denials++
	if s.cooldownN > 0 && c.denials >= s.cooldownN {
		c.cooldownUntil = now.Add(s.cooldownFor)
		c.denials = 0
		return int(s.cooldownFor.Seconds())
	}
	return 1
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
	for key, c := range s.cooldowns {
		if c.lastSeen.Before(cutoff) && !c.cooldownUntil.After(cutoff) {
			delete(s.cooldowns, key)
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
