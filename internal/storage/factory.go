// Package storage provides factory functions for creating storage services.
package storage

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/noruj-official/full-stack-go-template/internal/config"
)

// NewService creates a storage service based on the configuration.
func NewService(cfg *config.Config, db *pgxpool.Pool) (Service, error) {
	switch cfg.Storage.Type {
	case "database":
		return NewDatabaseStorage(db), nil
	case "s3":
		// TODO: Implement S3 storage
		return nil, fmt.Errorf("S3 storage not yet implemented. Please use 'database' storage for now")
	default:
		return nil, fmt.Errorf("unknown storage type: %s. Valid options: database, s3", cfg.Storage.Type)
	}
}
