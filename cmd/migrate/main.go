package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vyshakhp/utm-backend/internal/config"
	"github.com/vyshakhp/utm-backend/internal/database"
)

func main() {
	var (
		up       bool
		down     bool
		steps    int
		showHelp bool
	)

	flag.BoolVar(&up, "up", false, "Run all pending migrations")
	flag.BoolVar(&down, "down", false, "Rollback migrations")
	flag.IntVar(&steps, "steps", 1, "Number of migration steps to rollback (use with -down)")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.Parse()

	if showHelp || (!up && !down) {
		fmt.Println("Database Migration Tool")
		fmt.Println("\nUsage:")
		fmt.Println("  go run cmd/migrate/main.go -up              # Run all pending migrations")
		fmt.Println("  go run cmd/migrate/main.go -down -steps=1   # Rollback 1 migration")
		fmt.Println("  go run cmd/migrate/main.go -down -steps=2   # Rollback 2 migrations")
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if up {
		fmt.Println("Running migrations...")
		if err := database.RunMigrations(cfg); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migrations completed successfully")
	}

	if down {
		fmt.Printf("Rolling back %d migration(s)...\n", steps)
		if err := database.RollbackMigration(cfg, steps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✓ Rollback completed successfully")
	}
}
