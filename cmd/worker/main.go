package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vyshakhp/utm-backend/internal/config"
	"github.com/vyshakhp/utm-backend/internal/database"
	"github.com/vyshakhp/utm-backend/internal/jobs"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting UTM Backend Worker",
		zap.String("env", cfg.App.Env),
	)

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Database connection established")

	// Initialize worker
	worker, err := jobs.NewWorker(cfg, db, logger)
	if err != nil {
		logger.Fatal("Failed to initialize worker", zap.Error(err))
	}

	// Handle graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("Shutting down worker...")
		worker.Shutdown()
	}()

	// Start worker (blocking)
	logger.Info("Worker started, processing jobs...")
	if err := worker.Start(); err != nil {
		logger.Fatal("Worker error", zap.Error(err))
	}

	logger.Info("Worker exited")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.App.Env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
