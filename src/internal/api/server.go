package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	router     *mux.Router
	handler    *Handler
	httpServer *http.Server
	port       int
}

// NewServer creates a new API server
func NewServer(port int) *Server {
	router := mux.NewRouter()
	handler := NewHandler()

	// Register routes
	handler.RegisterRoutes(router)

	return &Server{
		router:  router,
		handler: handler,
		port:    port,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🚀 API server starting on port %d\n", s.port)
	fmt.Printf("📡 Endpoints:\n")
	fmt.Printf("   POST   http://localhost:%d/api/generate\n", s.port)
	fmt.Printf("   GET    http://localhost:%d/api/jobs/{job_id}\n", s.port)
	fmt.Printf("   GET    http://localhost:%d/api/jobs\n", s.port)
	fmt.Printf("   GET    http://localhost:%d/api/health\n", s.port)
	fmt.Println()

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
