package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/app/handlers"
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

	// Check if users table exists and has the correct schema
	var tableExists bool
	err = db.DB.QueryRow(
		`SELECT COUNT(*) > 0 FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = 'users'`).Scan(&tableExists)

	if err != nil {
		log.Fatalf("Error checking if users table exists: %v", err)
	}

	// Only run migrations if users table doesn't exist
	if !tableExists {
		// Run database migrations
		if err := db.RunMigrations(); err != nil {
			log.Fatalf("Error running database migrations: %v", err)
		}
	} else {
		log.Println("Users table already exists, skipping migrations")
	}

	// Initialize router with middleware
	router := setupRouter(db)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	// API v1 routes are registered in handlers.SetupRoutes

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
func setupRouter(db *database.Database) *gin.Engine {
	// Create a new Gin router with default middleware
	router := gin.New()

	// Add middleware
	handlers.SetupRoutes(router, db)

	return router
}
