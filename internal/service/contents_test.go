package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

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
