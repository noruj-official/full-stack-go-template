package domain

import (
	"time"

	"github.com/google/uuid"
)

// Media represents a stored file/image in the system.
type Media struct {
	ID          uuid.UUID  `json:"id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"` // Uploader
	Filename    string     `json:"filename"`
	Data        []byte     `json:"-"` // Binary data, not exposed in JSON
	ContentType string     `json:"content_type"`
	SizeBytes   int        `json:"size_bytes"`
	AltText     string     `json:"alt_text"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateMediaInput represents input for creating a new media item.
type CreateMediaInput struct {
	UserID      *uuid.UUID
	Filename    string
	Data        []byte
	ContentType string
	SizeBytes   int
	AltText     string
}

func (i *CreateMediaInput) Validate() error {
	if len(i.Data) == 0 {
		return ErrValidation{Field: "data", Message: "file data is required"}
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
