package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/J0es1ick/test-assignment/internal/balancer"
	"github.com/J0es1ick/test-assignment/internal/ratelimit"
)

type Server struct {
	server   *http.Server
	balancer *balancer.LoadBalancer
	limiter  *ratelimit.TokenBucketLimiter
}

func NewServer(port string, balancer *balancer.LoadBalancer, limiter *ratelimit.TokenBucketLimiter) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", balancer)

	handler := ratelimit.RateLimitMiddleware(
		limiter,
		func(r *http.Request) string {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			return ip
		},
	)(mux)

	return &Server{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
		balancer: balancer,
		limiter:  limiter,	
	}
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server")
	return s.server.Shutdown(ctx)
}