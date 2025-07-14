package database

import (
	"fmt"
	"os"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/app/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase() (*Database, error) {
	// Get database configuration
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	fmt.Printf("Connecting to database: %s@tcp(%s:%s)/%s\n", dbUser, dbHost, dbPort, dbName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	// Connect to MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.CrawlResult{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Println("Database migration completed successfully")

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("error getting database instance: %w", err)
	}

	return sqlDB.Close()
}

// Create wraps the GORM Create method
func (d *Database) Create(value interface{}) *gorm.DB {
	return d.DB.Create(value)
}

// First wraps the GORM First method
func (d *Database) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return d.DB.First(dest, conds...)
}

// Where wraps the GORM Where method
func (d *Database) Where(query interface{}, args ...interface{}) *gorm.DB {
	return d.DB.Where(query, args...)
}

// Exec wraps the GORM Exec method
func (d *Database) Exec(sql string, values ...interface{}) *gorm.DB {
	return d.DB.Exec(sql, values...)
}


