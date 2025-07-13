package database

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version string
	Up      string
}

// RunMigrations runs all pending database migrations
func (db *Database) RunMigrations() error {
	// First, ensure the database exists
	_, err := db.DB.Exec("CREATE DATABASE IF NOT EXISTS `url-analyzer`")
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}

	// Select the database
	_, err = db.DB.Exec("USE `url-analyzer`")
	if err != nil {
		return fmt.Errorf("error selecting database: %w", err)
	}

	// Create migrations table if it doesn't exist
	_, err = db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Start a transaction for all migrations
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	// Get applied migrations
	rows, err := tx.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error querying applied migrations: %w", err)
	}

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			tx.Rollback()
			return fmt.Errorf("error scanning migration version: %w", err)
		}
		applied[version] = true
	}
	rows.Close()

	// Read migration files
	migrationFiles, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error reading migration files: %w", err)
	}

	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, file := range migrationFiles {
		// Get just the base filename without path
		baseName := filepath.Base(file)
		// Remove .sql extension to get version
		version := strings.TrimSuffix(baseName, ".sql")
		// Extract numeric ID from filename (e.g., 1 from "1.sql")
		parts := strings.Split(version, "_")
		if len(parts) == 0 {
			tx.Rollback()
			return fmt.Errorf("invalid migration filename format: %s (should start with a number)", version)
		}
		
		id := 0
		_, err := fmt.Sscanf(parts[0], "%d", &id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("invalid migration filename format: %s (should start with a number)", version)
		}

		// Skip if already applied
		if applied[version] {
			continue
		}

		// Read migration SQL
		migrationSQL, err := fs.ReadFile(migrationsFS, file)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error reading migration %s: %w", file, err)
		}

		// Split SQL into individual statements and execute them one by one
		statements := strings.Split(string(migrationSQL), ";")
		for _, stmt := range statements {
			// Skip empty statements and comments
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}

			// Execute the statement
			log.Printf("Executing statement: %s\n", strings.SplitN(stmt, "\n", 2)[0] + "...")
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("error executing statement in migration %s: %w\nStatement: %s", version, err, stmt)
			}
		}

		// Record migration - use the full version string
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?) ON DUPLICATE KEY UPDATE version = VALUES(version)", version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error recording migration %s: %w", version, err)
		}
	}

	return tx.Commit()
}
