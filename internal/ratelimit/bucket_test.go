package ratelimit_test

import (
	"testing"
	"time"

	"github.com/J0es1ick/test-assignment/internal/ratelimit"
	"github.com/stretchr/testify/assert"
)

func TestTokenBucket(t *testing.T) {
	tb := ratelimit.NewTokenBucket(10, time.Second)

	t.Run("should allow requests when tokens available", func(t *testing.T) {
		allowed := 0
		for i := 0; i < 10; i++ {
			if tb.Take() {
				allowed++
			}
		}
		assert.Equal(t, 10, allowed)
	})

	t.Run("should reject when no tokens left", func(t *testing.T) {
		assert.False(t, tb.Take())
	})

	t.Run("should refill tokens after rate period", func(t *testing.T) {
		time.Sleep(1200 * time.Millisecond)
		assert.True(t, tb.Take())
	})
}
