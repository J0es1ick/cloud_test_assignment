package balancer

import (
	"context"
	"net"
	"net/url"
	"time"
)

type HealthChecker struct {
	pool     *BackendPool
	interval time.Duration	
}

func NewHealthCheck(pool *BackendPool, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		pool:     pool,
		interval: interval,
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
	for _, b := range hc.pool.backends {
		status := isBackendAlive(b.URL)
		b.SetAlive(status)
	}
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}