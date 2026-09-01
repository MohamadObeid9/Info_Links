package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"infolinks-backend/internal/repository"

	"golang.org/x/sync/singleflight"
)

const contentCacheTTL = 60 * time.Second

type ContentService struct {
	repo repository.ContentRepository
	ttl  time.Duration
	now  func() time.Time

	mu      sync.RWMutex
	cached  []byte
	expires time.Time
	gen     uint64
	group   singleflight.Group
}

func NewContentService(repo repository.ContentRepository) *ContentService {
	return &ContentService{
		repo: repo,
		ttl:  contentCacheTTL,
		now:  time.Now,
	}
}

// Get returns the public navigation JSON, serving a process-local copy when
// it is still within TTL. Concurrent misses share one Postgres round-trip.
func (c *ContentService) Get(ctx context.Context) ([]byte, error) {
	if b, ok := c.snapshot(); ok {
		return b, nil
	}

	v, err, _ := c.group.Do("content", func() (any, error) {
		if b, ok := c.snapshot(); ok {
			return b, nil
		}

		c.mu.RLock()
		gen := c.gen
		c.mu.RUnlock()

		result, err := c.repo.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("get content: %w", err)
		}
		cloned := append([]byte(nil), result...)

		c.mu.Lock()
		if c.gen == gen {
			c.cached = cloned
			c.expires = c.now().Add(c.ttl)
		}
		c.mu.Unlock()
		return cloned, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

// GetUncached always hits Postgres and does not read or write the student cache.
func (c *ContentService) GetUncached(ctx context.Context) ([]byte, error) {
	result, err := c.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get content: %w", err)
	}
	return result, nil
}

// Invalidate drops the student cache so the next Get refills from Postgres.
func (c *ContentService) Invalidate() {
	c.mu.Lock()
	c.cached = nil
	c.expires = time.Time{}
	c.gen++
	c.mu.Unlock()
}

func (c *ContentService) snapshot() ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cached == nil || !c.now().Before(c.expires) {
		return nil, false
	}
	return c.cached, true
}
