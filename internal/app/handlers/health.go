package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	// Add any dependencies here (e.g., database, services)
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles the health check endpoint
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	})
}
