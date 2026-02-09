package main

import (
	"log"

	"github.com/dhanuprys/infrantery-backend-go/internal/config"
	"github.com/dhanuprys/infrantery-backend-go/internal/server"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.Environment)
	logger.Info().
		Str("log_level", cfg.LogLevel).
		Str("environment", cfg.Environment).
		Msg("Logger initialized")

	// Initialize and run server
	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize server")
	}

	if err := srv.Run(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run server")
	}
}
