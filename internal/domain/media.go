package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StorageProviderDatabase = "database"
	StorageProviderS3       = "s3"
	StorageProviderLocal    = "local_fs"
)

// Media represents a stored file/image in the system.
type Media struct {
	ID              uuid.UUID  `json:"id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"` // Uploader
	Filename        string     `json:"filename"`
	Data            []byte     `json:"-"` // Binary data, not exposed in JSON (only if StorageProvider=database)
	ContentType     string     `json:"content_type"`
	SizeBytes       int        `json:"size_bytes"`
	AltText         string     `json:"alt_text"`
	StorageProvider string     `json:"storage_provider"`     // database, s3, etc.
	FileKey         string     `json:"file_key,omitempty"`   // S3 key or file path
	PublicURL       string     `json:"public_url,omitempty"` // Direct URL if available
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateMediaInput represents input for creating a new media item.
type CreateMediaInput struct {
	UserID          *uuid.UUID
	Filename        string
	Data            []byte
	ContentType     string
	SizeBytes       int
	AltText         string
	StorageProvider string // Optional, defaults to "database"
}

func (i *CreateMediaInput) Validate() error {
	// If storing in DB, Data is required.
	// We default to DB if not specified.
	provider := i.StorageProvider
	if provider == "" {
		provider = StorageProviderDatabase
	}

	if provider == StorageProviderDatabase && len(i.Data) == 0 {
		return ErrValidation{Field: "data", Message: "file data is required for database storage"}
	}

	if i.ContentType == "" {
		return ErrValidation{Field: "content_type", Message: "content type is required"}
	}
	// Max size check could be done here or service layer (e.g. 10MB)
	if i.SizeBytes > 10*1024*1024 {
		return ErrValidation{Field: "size", Message: "file too large (max 10MB)"}
	}
	return nil
}
