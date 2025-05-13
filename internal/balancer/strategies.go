package balancer

type RoundRobinStrategy struct {
	counter uint64
}

func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

func (s *RoundRobinStrategy) GetNextPeer(pool *BackendPool) *Backend {
	return pool.GetNextAlive()
}