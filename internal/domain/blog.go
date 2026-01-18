package domain

import (
	"time"

	"github.com/google/uuid"
)

// Blog represents a blog post.
type Blog struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	AuthorID    uuid.UUID `json:"author_id"`
	Author      *User     `json:"author,omitempty"` // Populated in some queries
	IsPublished bool      `json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
}

// UpdateBlogInput represents input for updating a blog.
type UpdateBlogInput struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Excerpt     *string `json:"excerpt"`
	Slug        *string `json:"slug"`
	IsPublished *bool   `json:"is_published"`
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
