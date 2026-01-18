// Package config provides configuration management for the application.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"fmt"
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
	Auth     AuthConfig
	Email    EmailConfig
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	Secret string
}

// EmailConfig contains email service settings.
type EmailConfig struct {
	ResendAPIKey    string
	ResendFromEmail string
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  string
	WriteTimeout string
	IdleTimeout  string
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
	URL  string
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
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         port,
			ReadTimeout:  getEnv("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getEnv("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getEnv("SERVER_IDLE_TIMEOUT", "60s"),
		},
		Database: DatabaseConfig{
			URL: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				getEnv("POSTGRES_USER", "postgres"),
				getEnv("POSTGRES_PASSWORD", "postgres"),
				getEnv("POSTGRES_HOST", "localhost"),
				getEnv("POSTGRES_PORT", "5432"),
				getEnv("POSTGRES_DB", "app_db"),
				getEnv("POSTGRES_SSLMODE", "disable"),
			),
		},
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Name: getEnv("APP_NAME", "Full Stack Go Template"),
			Logo: getEnv("APP_LOGO", "/static/img/logo.svg"),
			URL:  getEnv("APP_URL", "http://localhost:3000"),
		},
		Storage: StorageConfig{
			Type:     getEnv("PROFILE_IMAGE_STORAGE", "database"),
			S3Bucket: getEnv("S3_BUCKET", ""),
			S3Region: getEnv("S3_REGION", "us-east-1"),
		},
		Auth: AuthConfig{
			Secret: getEnv("AUTH_SECRET", ""),
		},
		Email: EmailConfig{
			ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
			ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
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
