package router

import (
	"blog-backend/internal/handler"
	"blog-backend/internal/middleware"
	"blog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Server static files
	service.ServerStatic(router)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	api := router.Group("/api/v1")
	{
		authHandler := handler.NewAuthHandler()
		postHandler := handler.NewPostHandler()
		categoryHandler := handler.NewCategoryHandler()
		tagHandler := handler.NewTagHandler()
		commentHandler := handler.NewCommentHandler()

		// Auth endpoints
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// Public post endpoints
		api.GET("/posts", postHandler.List)
		api.GET("/posts/slug/:slug", postHandler.GetBySlug)
		api.GET("/posts/:id", postHandler.GetByID)

		// Public category endpoints
		api.GET("/categories", categoryHandler.GetAll)
		api.GET("/categories/:id", categoryHandler.GetByID)

		// Public tag endpoints
		api.GET("/tags", tagHandler.GetAll)

		// Public comment endpoints
		api.GET("/comments/:post_id", commentHandler.GetByPost)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.Auth())
		{
			// Post management
			protected.POST("/posts", postHandler.Create)
			protected.PUT("/post/:id", postHandler.Update)
			protected.DELETE("/posts/:id", postHandler.Delete)

			// Comment management
			protected.POST("/comments", commentHandler.Create)

			// Upload management
			uploadHandler := handler.NewUploadHandler()
			protected.POST("/upload", uploadHandler.Upload)
			protected.GET("/uploads", uploadHandler.GetUserUploads)
			protected.DELETE("/upload/:id", uploadHandler.Delete)
		}

		// Admin only routes
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(), middleware.AdmiOnly())
		{
			admin.GET("/posts/all", postHandler.List) // List all posts including drafts

			// Category management
			admin.POST("/categories", categoryHandler.Create)
			admin.PUT("/categories/:id", categoryHandler.Update)
			admin.DELETE("/categories/:id", categoryHandler.Delete)

			// Tag management
			admin.POST("/tags", tagHandler.Create)
			admin.DELETE("/tags/:id", tagHandler.Delete)

			// Comment moderation
			admin.GET("/comments/pending", commentHandler.GetPending)
			admin.POST("/comments/:id/approve", commentHandler.Approve)
			admin.POST("/comments/:id/reject", commentHandler.Reject)

			// User management (can add later)
			// admin.GET("/users", userHandler.List)
			// admin.PUT("/users/:id/role", userHandler.UpdateRole)
		}
	}

	return router
}
