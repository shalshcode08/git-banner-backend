package server

import (
	"context"
	"net/http"
	"time"

	"github.com/somya/git-banner-backend/internal/config"
	"github.com/somya/git-banner-backend/internal/handler"
	"github.com/somya/git-banner-backend/internal/middleware"
)

// Server wraps http.Server with application configuration.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New builds the mux, registers routes, chains middleware, and configures timeouts.
func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)

	chain := middleware.Recovery(middleware.Logger(mux))

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      chain,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// Start begins listening for incoming connections.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully drains active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
