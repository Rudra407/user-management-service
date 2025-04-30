package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port int
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	JWT struct {
		Secret string
		Expiry int // in hours
	}
	Log struct {
		Level string
	}
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("Error loading .env file, using environment variables")
	}

	config := &Config{}

	// Server config
	if port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080")); err == nil {
		config.Server.Port = port
	} else {
		return nil, fmt.Errorf("invalid server port: %w", err)
	}

	// Database config
	config.Database.Host = getEnv("DB_HOST", "localhost")
	if port, err := strconv.Atoi(getEnv("DB_PORT", "5432")); err == nil {
		config.Database.Port = port
	} else {
		return nil, fmt.Errorf("invalid database port: %w", err)
	}
	config.Database.User = getEnv("DB_USER", "user")
	config.Database.Password = getEnv("DB_PASSWORD", "password")
	config.Database.Name = getEnv("DB_NAME", "testdb")
	config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	// JWT config
	config.JWT.Secret = getEnv("JWT_SECRET", "supersecretkey")
	if expiry, err := strconv.Atoi(getEnv("JWT_EXPIRY", "24")); err == nil {
		config.JWT.Expiry = expiry
	} else {
		return nil, fmt.Errorf("invalid JWT expiry: %w", err)
	}

	// Log config
	config.Log.Level = getEnv("LOG_LEVEL", "info")

	return config, nil
}

// DBConnectionString returns the PostgreSQL connection string
func (c *Config) DBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
