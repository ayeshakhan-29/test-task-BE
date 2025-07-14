package database

import (
	"fmt"

	"github.com/ayeshakhan-29/test-task-BE/internal/app/models"
)

// RunMigrations runs all pending database migrations using GORM's auto-migrate
func (db *Database) RunMigrations() error {
	// Auto-migrate the models - GORM will handle table creation and updates
	// based on the model definitions
	err := db.DB.AutoMigrate(
		&models.User{},
		&models.CrawlResult{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
