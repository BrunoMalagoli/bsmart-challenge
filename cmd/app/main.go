package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/config"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/middleware"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	db, err := config.NewDatabasePool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL successfully")

	// Initialize Gin router
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		models.RespondSuccess(c, http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/ready", func(c *gin.Context) {
		models.RespondSuccess(c, http.StatusOK, gin.H{"ready": true})
	})

	// TODO: Setup routes in internal/server/router.go
	// Future routes will be organized in internal/handlers

	// Start server
	addr := ":" + cfg.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
