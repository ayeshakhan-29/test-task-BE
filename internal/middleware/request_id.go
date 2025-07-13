package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader is the header key for the request ID
const RequestIDHeader = "X-Request-ID"

type contextKey string

const requestIDKey contextKey = "requestID"

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate a new one
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set the request ID in the context
		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Set the request ID in the response header
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID returns the request ID from the context
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Request.Context().Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	b := make([]byte, 12) // 96 bits
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
