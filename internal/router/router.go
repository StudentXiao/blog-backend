package router

import (
	"blog-backend/internal/handler"
	"blog-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	api := router.Group("/api/v1")
	{
		authHandler := handler.NewAuthHandler()
		postHandler := handler.NewPostHandler()

		// Auth endpoints
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// Public post endpoints
		api.GET("/posts", postHandler.List)
		api.GET("/posts/slug/:slug", postHandler.GetBySlug)
		api.GET("/posts/:id", postHandler.GetByID)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.Auth())
		{
			// Post management
			protected.POST("/posts", postHandler.Create)
			protected.PUT("/post/:id", postHandler.Update)
			protected.DELETE("/posts/:id", postHandler.Delete)
		}

		// Admin only routes
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(), middleware.AdmiOnly())
		{
			admin.GET("/posts/all", postHandler.List) // List all posts including drafts
		}
	}

	return router
}
