package handlers

import (
	"os"
	"strings"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/ayeshakhan-29/test-task-BE/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, db *database.Database) {
	// Configure CORS middleware
	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173" // Default value
	}

	// Split origins by comma and trim spaces
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}

	router.Use(cors.New(config))

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
			protected.GET("/analyzed-url/:id", crawlHandler.GetCrawlByID)
			protected.GET("/crawls", crawlHandler.ListCrawls)
			protected.DELETE("/delete/:id", crawlHandler.DeleteCrawl)
			protected.DELETE("/bulk-delete", crawlHandler.BulkDeleteCrawls)
		}
	}
}
