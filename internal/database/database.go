package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	DB *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	// First connect without specifying a database to create it if needed
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
		os.Getenv("DB_USER"),
		sqlEscapeString(os.Getenv("DB_PASSWORD")),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	tempDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Create the database if it doesn't exist
	_, err = tempDB.Exec("CREATE DATABASE IF NOT EXISTS `url-analyzer`")
	if err != nil {
		tempDB.Close()
		return nil, fmt.Errorf("error creating database: %w", err)
	}
	tempDB.Close()

	// Now connect to the specific database
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		sqlEscapeString(os.Getenv("DB_PASSWORD")),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		"url-analyzer",
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to database")
	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// sqlEscapeString escapes special characters in a string for SQL queries
func sqlEscapeString(value string) string {
	// In a real application, consider using parameterized queries instead of string escaping
	// This is a simplified version for demonstration
	var escape = map[rune]string{
		'\'': "''",
		'"':  "\"",
		'\b': "\\b",
		'\n': "\\n",
		'\r': "\\r",
		'\t': "\\t",
		'\\': "\\\\",
	}

	var result string
	for _, char := range value {
		if escaped, ok := escape[char]; ok {
			result += escaped
		} else {
			result += string(char)
		}
	}
	return result
}
