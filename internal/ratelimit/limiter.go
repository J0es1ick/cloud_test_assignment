package ratelimit

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type TokenBucketLimiter struct {
	defaultCapacity int
	defaultRate     time.Duration
	storage         Storage
	mux             sync.Mutex
}

func NewTokenBucketLimiter(defaultCapacity int, defaultRate time.Duration, storage Storage) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		defaultCapacity: defaultCapacity,
		defaultRate:     defaultRate,
		storage:         storage,
	}
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
    l.mux.Lock()
    defer l.mux.Unlock()

    var allowed bool
    err := l.storage.Update(ctx, key, func(b *TokenBucket) (*TokenBucket, error) {
        if b == nil {
            b = NewTokenBucket(l.defaultCapacity, l.defaultRate)
        }
        allowed = b.Take()
        return b, nil 
    })

    return allowed, err
}

func RateLimitMiddleware(limiter Limiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
            allowed, err := limiter.Allow(r.Context(), key) 
            if err != nil {
                log.Printf("Rate limiter error: %v", err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            if !allowed {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusTooManyRequests)
                w.Write([]byte(`{ "code": 429, "message": "Rate limit exceeded" }`))
                return
            }
			next.ServeHTTP(w, r)
		})
	}
}