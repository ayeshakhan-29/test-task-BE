package handlers

import (
	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/ayeshakhan-29/test-task-BE/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, db *database.Database) {
	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler(db)

	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", healthHandler.HealthCheck)

		// Authentication
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/signup", authHandler.Signup)
			authGroup.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := v1.Group("", middleware.AuthMiddleware())
		{
			crawlHandler := NewCrawlHandler(db)
			protected.POST("/crawl", crawlHandler.CrawlURL)
			protected.GET("/crawls", crawlHandler.ListCrawls)
			protected.DELETE("/delete/:id", crawlHandler.DeleteCrawl)
			protected.DELETE("/bulk-delete", crawlHandler.BulkDeleteCrawls)
		}
	}
}
