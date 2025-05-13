package balancer

import "net/http"

type LoadBalancer struct {
	pool     *BackendPool
	strategy Strategy
}

type Strategy interface {
	GetNextPeer(*BackendPool) *Backend
}

func NewLoadBalancer(pool *BackendPool, strategy Strategy) *LoadBalancer {
	return &LoadBalancer{
		pool:     pool,
		strategy: strategy,
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.strategy.GetNextPeer(lb.pool)
	if backend != nil {
		backend.ReverseProxy.ServeHTTP(w, r)
		return
	}
	
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}