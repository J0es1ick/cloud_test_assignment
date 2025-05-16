package balancer

import (
	"context"
	"net"
	"net/url"
	"sync"
	"time"
)

type HealthChecker struct {
	pool     *BackendPool
	interval time.Duration
	timeout  time.Duration
}

func NewHealthChecker(pool *BackendPool, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		pool:     pool,
		interval: interval,
		timeout:  2 * time.Second, 
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.healthCheck()
		case <-ctx.Done():
			return
		}
	}
}

func (hc *HealthChecker) healthCheck() {
	var wg sync.WaitGroup

	for _, b := range hc.pool.Backends {
		wg.Add(1)
		go func(backend *Backend) {
			defer wg.Done()
			status := hc.isBackendAlive(backend.URL)
			backend.SetAlive(status)
		}(b)
	}

	wg.Wait()
}

func (hc *HealthChecker) isBackendAlive(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, hc.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}