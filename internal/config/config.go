package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	AppVersion string
	ServerPort string
	Database   DatabaseConfig
	Server     ServerConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Set default values
	cfg := &Config{
		Environment: getEnv("ENV", "development"),
		AppVersion:  getEnv("APP_VERSION", "1.0.0"),
		ServerPort:  getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "3306"),
			User:         getEnv("DB_USER", "root"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "test_task"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			MaxIdleTime:  time.Duration(getEnvAsInt("DB_MAX_IDLE_TIME", 15)) * time.Minute,
		},
		Server: ServerConfig{
			Port:              getEnv("PORT", "8080"),
			ReadTimeout:       time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout:      time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 30)) * time.Second,
			IdleTimeout:       time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
			ReadHeaderTimeout: time.Duration(getEnvAsInt("SERVER_READ_HEADER_TIMEOUT", 5)) * time.Second,
		},
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// IsProduction returns true if the environment is set to production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is set to development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTesting returns true if the environment is set to testing
func (c *Config) IsTesting() bool {
	return c.Environment == "testing"
}
