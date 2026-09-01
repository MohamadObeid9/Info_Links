package service

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"infolinks-backend/internal/errs"
)

type fakeContentRepo struct {
	getCalls  int
	getResult []byte
	getErr    error
}

func (f *fakeContentRepo) Get(ctx context.Context) ([]byte, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResult, nil
}

func TestContentService_Get(t *testing.T) {
	sampleJSON := []byte(`{"programs":[],"years":[]}`)

	tests := []struct {
		name       string
		repoResult []byte
		repoErr    error
		wantResult []byte
		wantErr    error
		wantCalls  int
	}{
		{
			name:       "returns repo result",
			repoResult: sampleJSON,
			wantResult: sampleJSON,
			wantCalls:  1,
		},
		{
			name:      "wraps repo error",
			repoErr:   errs.ErrDatabaseDown,
			wantErr:   errs.ErrDatabaseDown,
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeContentRepo{getResult: tt.repoResult, getErr: tt.repoErr}
			svc := NewContentService(repo)

			got, err := svc.Get(context.Background())
			if repo.getCalls != tt.wantCalls {
				t.Fatalf("repo get calls = %d, want %d", repo.getCalls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Fatalf("got %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestContentService_Get_usesCacheUntilInvalidateOrTTL(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	repo := &fakeContentRepo{getResult: []byte(`{"ok":true}`)}
	svc := NewContentService(repo)
	svc.now = func() time.Time { return now }

	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if repo.getCalls != 1 {
		t.Fatalf("repo calls after two Gets = %d, want 1", repo.getCalls)
	}

	svc.Invalidate()
	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get after invalidate: %v", err)
	}
	if repo.getCalls != 2 {
		t.Fatalf("repo calls after invalidate = %d, want 2", repo.getCalls)
	}

	now = now.Add(contentCacheTTL + time.Second)
	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get after TTL: %v", err)
	}
	if repo.getCalls != 3 {
		t.Fatalf("repo calls after TTL = %d, want 3", repo.getCalls)
	}
}

func TestContentService_GetUncached_skipsCache(t *testing.T) {
	repo := &fakeContentRepo{getResult: []byte(`{"ok":true}`)}
	svc := NewContentService(repo)

	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := svc.GetUncached(context.Background()); err != nil {
		t.Fatalf("GetUncached: %v", err)
	}
	if repo.getCalls != 2 {
		t.Fatalf("repo calls = %d, want 2", repo.getCalls)
	}
	if _, err := svc.Get(context.Background()); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if repo.getCalls != 2 {
		t.Fatalf("cached Get after GetUncached hit repo again: calls = %d", repo.getCalls)
	}
}

func TestContentService_Get_singleflight(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	repo := &blockingContentRepo{
		fakeContentRepo: fakeContentRepo{getResult: []byte(`{"ok":true}`)},
		started:         started,
		release:         release,
	}
	svc := NewContentService(repo)

	errc := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := svc.Get(context.Background())
			errc <- err
		}()
	}
	<-started
	close(release)
	for range 2 {
		if err := <-errc; err != nil {
			t.Fatalf("Get: %v", err)
		}
	}
	if repo.getCalls != 1 {
		t.Fatalf("concurrent misses hit repo %d times, want 1", repo.getCalls)
	}
}

type blockingContentRepo struct {
	fakeContentRepo
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *blockingContentRepo) Get(ctx context.Context) ([]byte, error) {
	b.once.Do(func() { close(b.started) })
	<-b.release
	return b.fakeContentRepo.Get(ctx)
}
