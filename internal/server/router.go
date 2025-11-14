package server

import (
	"log"
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/handlers"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/middleware"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetupRouter(database *db.DB, jwtService *auth.JWTService, hub *websockets.Hub) *gin.Engine {
	// Create router
	r := gin.Default()

	// global middleware
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	h := handlers.NewHandler(database, jwtService, hub)

	r.GET("/health", func(c *gin.Context) {
		models.RespondSuccess(c, http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes - no authentication required
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", h.Register)
			authRoutes.POST("/login", h.Login)
		}

		// Protected routes - authentication required
		protected := api.Group("")
		protected.Use(middleware.RequireAuth(jwtService))
		{

			protected.GET("/products", h.ListProducts)
			protected.GET("/products/:id", h.GetProduct)
			protected.GET("/products/:id/history", h.GetProductHistory)

			protected.GET("/categories", h.ListCategories)
			protected.GET("/categories/:id", h.GetCategory)

			protected.GET("/search", h.Search)

			// Admin-only routes
			admin := protected.Group("")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.POST("/products", h.CreateProduct)
				admin.PUT("/products/:id", h.UpdateProduct)
				admin.DELETE("/products/:id", h.DeleteProduct)

				admin.POST("/categories", h.CreateCategory)
				admin.PUT("/categories/:id", h.UpdateCategory)
				admin.DELETE("/categories/:id", h.DeleteCategory)
			}
		}
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade to websocket: %v", err)
			return
		}

		client := websockets.NewClient(hub, conn)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()
	})

	return r
}
