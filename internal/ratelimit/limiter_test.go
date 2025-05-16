package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/J0es1ick/test-assignment/internal/ratelimit"
	"github.com/stretchr/testify/assert"
)

func TestTokenBucketLimiter(t *testing.T) {
	t.Run("should allow when tokens available", func(t *testing.T) {
		mockStorage := &mockStorage{
			bucket: ratelimit.NewTokenBucket(10, time.Second),
		}

		limiter := ratelimit.NewTokenBucketLimiter(10, time.Second, mockStorage)
		allowed, err := limiter.Allow(context.Background(), "test")

		assert.True(t, allowed)
		assert.NoError(t, err)
	})

	t.Run("should reject when rate limit exceeded", func(t *testing.T) {
		mockStorage := &mockStorage{
			bucket: ratelimit.NewTokenBucket(1, time.Minute),
		}

		limiter := ratelimit.NewTokenBucketLimiter(1, time.Minute, mockStorage)
		
		allowed, err := limiter.Allow(context.Background(), "test")
		assert.True(t, allowed)
		assert.NoError(t, err)
		
		allowed, err = limiter.Allow(context.Background(), "test")
		assert.False(t, allowed)
		assert.NoError(t, err)
	})
}

type mockStorage struct {
	bucket *ratelimit.TokenBucket
}

func (m *mockStorage) Get(ctx context.Context, key string) (*ratelimit.TokenBucket, bool, error) {
	if m.bucket == nil {
		return nil, false, nil
	}
	return m.bucket, true, nil
}

func (m *mockStorage) Set(ctx context.Context, key string, bucket *ratelimit.TokenBucket) error {
	m.bucket = bucket
	return nil
}

func (m *mockStorage) Update(ctx context.Context, key string, updateFunc func(*ratelimit.TokenBucket) (*ratelimit.TokenBucket, error)) error {
	newBucket, err := updateFunc(m.bucket)
	if err != nil {
		return err
	}
	m.bucket = newBucket
	return nil
}

