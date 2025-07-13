package handlers

import "github.com/gin-gonic/gin"

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine) {
	// Initialize handlers
	healthHandler := NewHealthHandler()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", healthHandler.HealthCheck)

		// Add more routes here
	}
}
