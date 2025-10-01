package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	router *chi.Mux
	server *http.Server
}

func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = &Config{
			Port:         "8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}

	r := chi.NewRouter()
	return &Server{
		router: r,
		server: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      r,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}, nil
}

// Router returns the Chi router for route registration
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Start starts the HTTP server
func (s *Server) start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Start() error {
	errChan := make(chan error, 1)
	go func() {
		if err := s.start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf(
			"server failed to start: %w", err)
	case <-quit:
		log.Println("Shutting down server...")
		// Let user handle timeout if they want
		return s.server.Shutdown(context.Background())
	}
}
