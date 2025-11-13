package server

import (
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/handlers"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/middleware"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

func SetupRouter(database *db.DB, jwtService *auth.JWTService) *gin.Engine {
	// Create router
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	h := handlers.NewHandler(database, jwtService)

	r.GET("/health", func(c *gin.Context) {
		models.RespondSuccess(c, http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Public routes - no authentication required
		public := api.Group("")
		{
			// Products - read only
			public.GET("/products", h.ListProducts)
			public.GET("/products/:id", h.GetProduct)
			public.GET("/products/:id/history", h.GetProductHistory)

			// Categories - read only
			public.GET("/categories", h.ListCategories)
			public.GET("/categories/:id", h.GetCategory)

			// Search
			public.GET("/search", h.Search)
		}

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

	return r
}
