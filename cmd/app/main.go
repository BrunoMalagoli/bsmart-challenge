package main

import (
	"fmt"
	"log"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/config"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/server"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
)

func main() {
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

	// Wrap pool in DB struct
	database := db.NewDB(pool)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// Initialize WebSocket hub
	hub := websockets.NewHub()
	go hub.Run() // Start hub in a goroutine

	// Setup router with all routes and middleware
	router := server.SetupRouter(database, jwtService, hub)

	// Start server
	addr := ":" + cfg.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
