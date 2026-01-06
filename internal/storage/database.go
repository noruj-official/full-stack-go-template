// Package storage provides database-based storage implementation.
package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseStorage implements the Service interface using PostgreSQL.
type DatabaseStorage struct {
	db *pgxpool.Pool
}

// NewDatabaseStorage creates a new database storage service.
func NewDatabaseStorage(db *pgxpool.Pool) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

// StoreProfileImage stores a profile image in the database.
func (s *DatabaseStorage) StoreProfileImage(ctx context.Context, userID uuid.UUID, imageData []byte, contentType string, size int) error {
	query := `
		UPDATE users 
		SET profile_image = $1, 
		    profile_image_type = $2, 
		    profile_image_size = $3,
		    updated_at = NOW()
		WHERE id = $4
	`

	result, err := s.db.Exec(ctx, query, imageData, contentType, size, userID)
	if err != nil {
		return fmt.Errorf("failed to store profile image: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetProfileImage retrieves a profile image from the database.
func (s *DatabaseStorage) GetProfileImage(ctx context.Context, userID uuid.UUID) ([]byte, string, error) {
	query := `
		SELECT profile_image, profile_image_type 
		FROM users 
		WHERE id = $1 AND profile_image IS NOT NULL
	`

	var imageData []byte
	var contentType string

	err := s.db.QueryRow(ctx, query, userID).Scan(&imageData, &contentType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get profile image: %w", err)
	}

	return imageData, contentType, nil
}

// DeleteProfileImage removes a profile image from the database.
func (s *DatabaseStorage) DeleteProfileImage(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET profile_image = NULL, 
		    profile_image_type = NULL, 
		    profile_image_size = 0,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete profile image: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
