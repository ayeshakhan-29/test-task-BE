package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func dropAllTables(db *sql.DB) error {
	tables := []string{
		"users",
		"schema_migrations",
	}

	// Disable foreign key checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("error disabling foreign key checks: %v", err)
	}

	// Drop tables
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			// Re-enable foreign key checks even if there's an error
			db.Exec("SET FOREIGN_KEY_CHECKS = 1")
			return fmt.Errorf("error dropping table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("error re-enabling foreign key checks: %v", err)
	}

	return nil
}

func main() {
	// Get database connection string from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "root:@tcp(localhost:3306)/url-analyzer"
	}

	// Connect to database
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Drop all existing tables to start fresh
	if err := dropAllTables(db); err != nil {
		log.Fatalf("Error dropping tables: %v", err)
	}

	// Create schema_migrations table
	_, err = db.Exec(`
		CREATE TABLE schema_migrations (
			version VARCHAR(255) NOT NULL PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		log.Fatalf("Error creating schema_migrations table: %v", err)
	}

	// Check which migrations have already been applied
	var appliedMigrations []string
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var version string
			if err := rows.Scan(&version); err == nil {
				appliedMigrations = append(appliedMigrations, version)
			}
		}
	}

	// Define migrations in order
	migrations := []struct {
		version string
		query   string
	}{
		{
			version: "1",
			query: `
				CREATE TABLE IF NOT EXISTS users (
					id INT AUTO_INCREMENT PRIMARY KEY,
					username VARCHAR(50) NOT NULL UNIQUE,
					email VARCHAR(100) NOT NULL UNIQUE,
					password_hash VARCHAR(255) NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
			`,
		},
		{
			version: "2",
			query: `
				ALTER TABLE users
				ADD COLUMN full_name VARCHAR(100) NOT NULL AFTER id;
			`,
		},
		{
			version: "3",
			query: `
				ALTER TABLE users
				DROP COLUMN username;
			`,
		},
		{
			version: "4",
			query: `
				ALTER TABLE users
				MODIFY COLUMN full_name VARCHAR(100) NOT NULL AFTER id;
			`,
		},
	}

	// Apply migrations
	for _, migration := range migrations {
		// Skip if already applied
		alreadyApplied := false
		for _, applied := range appliedMigrations {
			if applied == migration.version {
				alreadyApplied = true
				break
			}
		}

		if !alreadyApplied {
			// Start transaction
			tx, err := db.Begin()
			if err != nil {
				log.Fatalf("Error starting transaction: %v", err)
			}

			// Execute migration
			_, err = tx.Exec(migration.query)
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error applying migration %s: %v", migration.version, err)
			}

			// Record migration
			_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.version)
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error recording migration %s: %v", migration.version, err)
			}

			// Commit transaction
			err = tx.Commit()
			if err != nil {
				log.Fatalf("Error committing migration %s: %v", migration.version, err)
			}

			fmt.Printf("Applied migration %s\n", migration.version)
		} else {
			fmt.Printf("Migration %s already applied\n", migration.version)
		}
	}

	fmt.Println("Migrations completed successfully")
}
