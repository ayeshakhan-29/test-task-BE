package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/logger"
	"github.com/gin-gonic/gin"
)

// bodyLogWriter is a custom response writer that captures the response body
// for logging purposes.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body for logging.
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware logs HTTP requests and responses.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Record the start time
		start := time.Now()

		// Capture the request body if it exists
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture the response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process the request
		c.Next()

		// Calculate the request duration
		duration := time.Since(start)

		// Skip logging for static files
		if c.Writer.Status() == 304 {
			return
		}

		// Log the request and response details
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"request_id": GetRequestID(c),
		}

		// Add request body if it exists and is not too large
		if len(requestBody) > 0 && len(requestBody) < 1024 {
			fields["request_body"] = string(requestBody)
		}

		// Add response body if it exists and is not too large
		if blw.body.Len() > 0 && blw.body.Len() < 1024 {
			fields["response_body"] = blw.body.String()
		}

		// Log the request
		if c.Writer.Status() >= 400 {
			logger.Error("HTTP Request Error", fields)
		} else {
			logger.Info("HTTP Request", fields)
		}
	}
}
