package balancer

import "sync"

type Strategy interface {
	GetNextPeer(*BackendPool) *Backend
}

type RoundRobinStrategy struct {
	counter uint64
	mux      sync.Mutex
}

func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

func (s *RoundRobinStrategy) GetNextPeer(pool *BackendPool) *Backend {
	s.mux.Lock()
	defer s.mux.Unlock()

	backends := pool.GetBackends()
	
	aliveBackends := make([]*Backend, 0, len(backends))
	for _, b := range backends {
		if b.IsAlive() {
			aliveBackends = append(aliveBackends, b)
		}
	}

	if len(aliveBackends) == 0 {
		return nil
	}

	next := int(s.counter % uint64(len(aliveBackends)))
	s.counter++
	return aliveBackends[next]
}