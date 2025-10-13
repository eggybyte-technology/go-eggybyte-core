package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

func main() {
	// Load configuration from environment variables
	cfg := &config.Config{
		ServiceName: "basic-example",
		Port:        8080,
		LogLevel:    "info",
		LogFormat:   "console",
	}
	config.MustReadFromEnv(cfg)

	// Create a simple HTTP server
	httpServer := &SimpleHTTPServer{
		port: cfg.Port,
	}

	// Bootstrap the service with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := core.Bootstrap(ctx, cfg, httpServer); err != nil {
		log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
	}

	log.Info("Service started successfully", log.Field{Key: "port", Value: cfg.Port})
	<-ctx.Done()
	log.Info("Service shutting down...")
}

// SimpleHTTPServer implements the service.Service interface
type SimpleHTTPServer struct {
	port int
}

func (s *SimpleHTTPServer) Start(ctx context.Context) error {
	log.Info("Starting HTTP server", log.Field{Key: "port", Value: s.port})

	// Simulate server startup
	time.Sleep(100 * time.Millisecond)

	log.Info("HTTP server started successfully")
	return nil
}

func (s *SimpleHTTPServer) Stop(ctx context.Context) error {
	log.Info("Stopping HTTP server")

	// Simulate graceful shutdown
	time.Sleep(50 * time.Millisecond)

	log.Info("HTTP server stopped")
	return nil
}

func (s *SimpleHTTPServer) Name() string {
	return "simple-http-server"
}
