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

// @title           Bsmart Backend API
// @version         1.0
// @description     RESTful API con soporte de WebSocket en tiempo real para gestión de productos y categorías. Sistema completo con autenticación JWT, control de acceso basado en roles, búsqueda full-text y seguimiento automático de historial de productos.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Bsmart API Support
// @contact.email  bruno.malagoli@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Autenticación JWT. Formato: "Bearer {token}". Obtén un token desde /api/auth/login o /api/auth/register

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
