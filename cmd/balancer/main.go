package main

import (
	"context"
	"log"
	"time"

	"github.com/J0es1ick/test-assignment/internal/balancer"
	"github.com/J0es1ick/test-assignment/internal/config"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	backendPool := balancer.NewBackendPool(cfg.Backends)
	lb := balancer.NewLoadBalancer(backendPool, balancer.NewRoundRobinStrategy())

	healthChecker := balancer.NewHealthCheck(backendPool, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go healthChecker.Start(ctx)
}