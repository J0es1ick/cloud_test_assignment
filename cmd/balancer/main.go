package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/J0es1ick/test-assignment/internal/balancer"
	"github.com/J0es1ick/test-assignment/internal/config"
	"github.com/J0es1ick/test-assignment/internal/ratelimit"
	"github.com/J0es1ick/test-assignment/internal/server"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGHUP)
		for {
			<-sigchan
			if err := config.ReloadConfig(); err != nil {
				log.Printf("Config reload failed: %v", err)
			}
		}
	}()

	backendPool := balancer.NewBackendPool(cfg.Backends)
	lb := balancer.NewLoadBalancer(backendPool, balancer.NewRoundRobinStrategy())

	healthChecker := balancer.NewHealthChecker(backendPool, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go healthChecker.Start(ctx)

	db, err := ratelimit.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Init(ctx); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	cleanupCtx := context.Background()
    db.StartCleanupWorker(cleanupCtx, 6*time.Hour)
	
	limiter := ratelimit.NewTokenBucketLimiter(cfg.Ratelimit.DefaultCapacity, cfg.Ratelimit.DefaultRate, db)

	server := server.NewServer(cfg.Server.Port, lb, limiter)
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			quit <- syscall.SIGTERM
		}
	}()

	<- quit
	log.Println("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}