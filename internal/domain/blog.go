package domain

import (
	"time"

	"github.com/google/uuid"
)

// Blog represents a blog post.
type Blog struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	Excerpt     string     `json:"excerpt"`
	AuthorID    uuid.UUID  `json:"author_id"`
	Author      *User      `json:"author,omitempty"` // Populated in some queries
	IsPublished bool       `json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`

	// Cover Image
	CoverMediaID *uuid.UUID `json:"cover_media_id,omitempty"`
	CoverMedia   *Media     `json:"-"`

	// SEO Metadata
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	MetaKeywords    string `json:"meta_keywords,omitempty"`
	OGImage         []byte `json:"-"` // OpenGraph image
	OGImageType     string `json:"og_image_type,omitempty"`
	OGImageSize     int    `json:"og_image_size,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BlogFilter defines criteria for listing blogs.
type BlogFilter struct {
	IsPublished *bool
	AuthorID    *uuid.UUID
	Limit       int
	Offset      int
}

// CreateBlogInput represents input for creating a blog.
type CreateBlogInput struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Excerpt     string `json:"excerpt"`
	IsPublished bool   `json:"is_published"`

	// SEO Metadata (optional)
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`

	// Cover Image (Upload)
	CoverImage []byte `json:"-"`
}

// UpdateBlog Input represents input for updating a blog.
type UpdateBlogInput struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Excerpt     *string `json:"excerpt"`
	Slug        *string `json:"slug"`
	IsPublished *bool   `json:"is_published"`

	// SEO Metadata (optional)
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
	MetaKeywords    *string `json:"meta_keywords"`

	// Cover Image (Upload)
	CoverImage []byte `json:"-"`
	// Flag to remove existing cover image
	RemoveCoverImage bool `json:"-"`
	// Set cover to existing media ID (from gallery)
	CoverMediaID *uuid.UUID `json:"-"`
	// Metadata handled by Media service creation
}

func (i *CreateBlogInput) Validate() error {
	if i.Title == "" {
		return ErrValidation{Field: "title", Message: "title is required"}
	}
	if i.Content == "" {
		return ErrValidation{Field: "content", Message: "content is required"}
	}
	return nil
}
