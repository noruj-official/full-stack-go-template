// Package config provides configuration management for the application.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
	Storage  StorageConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	URL string
}

// AppConfig contains general application settings.
type AppConfig struct {
	Env  string
	Name string
	Logo string
}

// StorageConfig contains file/image storage settings.
type StorageConfig struct {
	// Type determines where profile images are stored: "database" or "s3"
	Type string
	// S3Bucket is the S3 bucket name (only used when Type is "s3")
	S3Bucket string
	// S3Region is the AWS region (only used when Type is "s3")
	S3Region string
}

// Load reads configuration from environment variables.
// It attempts to load a .env file first, but continues if not found.
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("SERVER_PORT", "3000"))
	if err != nil {
		port = 3000
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: port,
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable"),
		},
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Name: getEnv("APP_NAME", "Full Stack Go Template"),
			Logo: getEnv("APP_LOGO", "/static/img/logo.svg"),
		},
		Storage: StorageConfig{
			Type:     getEnv("PROFILE_IMAGE_STORAGE", "database"),
			S3Bucket: getEnv("S3_BUCKET", ""),
			S3Region: getEnv("S3_REGION", "us-east-1"),
		},
	}, nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
