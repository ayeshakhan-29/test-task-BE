package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/config"
	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/ayeshakhan-29/test-task-BE/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	// Initialize logger
	logger.InitLogger()

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database connection
	db, err := database.NewDatabase()
	if err != nil {
		logger.Fatalf("Error initializing database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatalf("Error closing database connection: %v", err)
		}
	}()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		logger.Fatalf("Error running database migrations: %v", err)
	}

	// Initialize router with middleware
	router := setupRouter()

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"version": "1.0.0",
			})
		})
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server is running on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly")
}

// setupRouter initializes and configures the Gin router with middleware and routes
func setupRouter() *gin.Engine {
	// Create a new Gin router
	router := gin.New()

	// Add built-in Gin middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}
