package domain

import (
	"time"

	"github.com/google/uuid"
)

// BlogImage represents an image in a blog's gallery.
type BlogImage struct {
	ID        uuid.UUID `json:"id"`
	BlogID    uuid.UUID `json:"blog_id"`
	MediaID   uuid.UUID `json:"media_id"`
	Media     *Media    `json:"-"`
	AltText   string    `json:"alt_text"`
	Caption   string    `json:"caption"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateBlogImageInput represents input for creating a blog image.
type CreateBlogImageInput struct {
	BlogID      uuid.UUID `json:"blog_id"`
	ImageData   []byte    `json:"-"`
	ContentType string    `json:"-"` // Needed for media creation
	// Size calculated from data
	AltText  string `json:"alt_text"`
	Caption  string `json:"caption"`
	Position int    `json:"position"`
}

// UpdateBlogImageInput represents input for updating a blog image.
type UpdateBlogImageInput struct {
	AltText  *string `json:"alt_text"`
	Caption  *string `json:"caption"`
	Position *int    `json:"position"`
}

func (i *CreateBlogImageInput) Validate() error {
	if i.BlogID == uuid.Nil {
		return ErrValidation{Field: "blog_id", Message: "blog_id is required"}
	}
	if len(i.ImageData) == 0 {
		return ErrValidation{Field: "image_data", Message: "image_data is required"}
	}
	if i.ContentType == "" {
		return ErrValidation{Field: "content_type", Message: "content_type is required"}
	}
	return nil
}
