package main

import (
	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/core"
	"github.com/eggybyte-technology/go-eggybyte-core/log"

	// Import repositories for auto-registration
	_ "github.com/eggybyte-technology/demo-platform/backend/services/auth/internal/repositories"
)

func main() {
	// Load configuration from environment
	cfg := &config.Config{}
	config.MustReadFromEnv(cfg)

	// Bootstrap service with core infrastructure
	if err := core.Bootstrap(cfg); err != nil {
		log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
	}
}
