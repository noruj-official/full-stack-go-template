// Package storage defines the interface for file/image storage operations.
package storage

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for storing and retrieving files.
// This abstraction allows for different storage backends (database, S3, filesystem, etc.)
type Service interface {
	// StoreProfileImage stores a profile image for a user.
	// Returns an error if the storage operation fails.
	StoreProfileImage(ctx context.Context, userID uuid.UUID, imageData []byte, contentType string, size int) error

	// GetProfileImage retrieves a profile image for a user.
	// Returns the image data and content type, or an error if not found.
	GetProfileImage(ctx context.Context, userID uuid.UUID) (imageData []byte, contentType string, err error)

	// DeleteProfileImage removes a profile image for a user.
	// Returns an error if the deletion fails.
	DeleteProfileImage(ctx context.Context, userID uuid.UUID) error
}
