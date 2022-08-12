package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/oramaz/lb/internal/pool"
)

// Server describes the load-balancer server
type Server struct {
	Port string
	Pool *pool.ConnPool
}

// New returns an Server object
func New(port int, pool *pool.ConnPool) *Server {
	return &Server{
		Port: fmt.Sprintf(":%d", port),
		Pool: pool,
	}
}

// Start runs load-balancer server
func (s *Server) Start() error {
	server := http.Server{
		Addr:    s.Port,
		Handler: http.HandlerFunc(s.lb),
	}

	// Run passive health check
	go s.healthCheck()
	// Run load logs
	go s.loadStatistics()

	log.Printf("Load-balancer service started on port %s\n", s.Port)
	return server.ListenAndServe()
}

// lb is a load-balancer handler
func (s *Server) lb(w http.ResponseWriter, r *http.Request) {
	peer := s.Pool.Next()
	if peer != nil {
		peer.Proxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// healthCheck calls pool's HealhCheck function every 45s
func (s *Server) healthCheck() {
	t := time.NewTicker(time.Second * 45)

	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			s.Pool.HealthCheck()
			log.Println("Health check completed.")
		default:
			continue
		}
	}
}

// loadStatistics calls pool's LoadStatistics function every 10s
func (s *Server) loadStatistics() {
	t := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-t.C:
			log.Println("Connection's load:")
			s.Pool.LoadStatistics()
		default:
			continue
		}
	}
}
