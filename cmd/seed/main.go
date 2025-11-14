package main

import (
	"fmt"
	"log"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/config"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/seed"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Bsmart Backend - Database Seeder")
	fmt.Println("===========================================")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	pool, err := config.NewDatabasePool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Connected to PostgreSQL successfully")
	fmt.Println()

	// Wrap pool in DB struct
	database := db.NewDB(pool)

	// Run seeders
	if err := seed.SeedAll(database); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("  Seeding completed successfully!")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("Test Users:")
	fmt.Println("  Admin:  admin@bsmart.com / admin123")
	fmt.Println("  Client: client@bsmart.com / client123")
	fmt.Println()
}
