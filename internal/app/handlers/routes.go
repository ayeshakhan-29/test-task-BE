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

		// Auth routes
		v1.POST("/signup", authHandler.Signup)
		v1.POST("/login", authHandler.Login)

		// Protected routes
		authRoutes := v1.Group("/")
		authRoutes.Use(middleware.AuthMiddleware())
		// Initialize the handler once
		crawlHandler := NewCrawlHandler(db)
		// Register the route with the full path
		authRoutes.POST("crawl", crawlHandler.CrawlURL)
	}
}
