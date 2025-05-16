package balancer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/J0es1ick/test-assignment/internal/balancer"
	"github.com/stretchr/testify/assert"
)

func TestHealthChecker(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer backend.Close()

	pool := balancer.NewBackendPool([]string{backend.URL})
	checker := balancer.NewHealthChecker(pool, 100*time.Millisecond)

	t.Run("should detect live backends", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go checker.Start(ctx)
		time.Sleep(150 * time.Millisecond)

		assert.True(t, pool.Backends[0].IsAlive())
	})

	t.Run("should detect dead backends", func(t *testing.T) {
		backend.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go checker.Start(ctx)
		time.Sleep(150 * time.Millisecond)

		assert.False(t, pool.Backends[0].IsAlive())
	})
}