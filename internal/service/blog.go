package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

type BlogRepository interface {
	Create(ctx context.Context, blog *domain.Blog) error
	Update(ctx context.Context, blog *domain.Blog) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Blog, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Blog, error)
	List(ctx context.Context, filter domain.BlogFilter) ([]*domain.Blog, int, error)
}

type BlogService struct {
	repo BlogRepository
}

func NewBlogService(repo BlogRepository) *BlogService {
	return &BlogService{repo: repo}
}

func (s *BlogService) Create(ctx context.Context, input domain.CreateBlogInput, authorID uuid.UUID) (*domain.Blog, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	slug := generateSlug(input.Title)
	// TODO: Handle slug collision if needed (e.g. append random string)

	now := time.Now()
	blog := &domain.Blog{
		ID:          uuid.New(),
		Title:       input.Title,
		Slug:        slug,
		Content:     sanitizeContent(input.Content),
		Excerpt:     input.Excerpt,
		AuthorID:    authorID,
		IsPublished: input.IsPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if input.IsPublished {
		blog.PublishedAt = &now
	}

	if err := s.repo.Create(ctx, blog); err != nil {
		return nil, err
	}

	return blog, nil
}

func (s *BlogService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateBlogInput) (*domain.Blog, error) {
	blog, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, fmt.Errorf("blog not found")
	}

	if input.Title != nil {
		blog.Title = *input.Title
		// If slug is not explicitly updated, maybe regenerate?
		// Usually we don't change slug automatically on update to preserve links.
	}
	if input.Slug != nil {
		blog.Slug = generateSlug(*input.Slug)
	}
	if input.Content != nil {
		blog.Content = sanitizeContent(*input.Content)
	}
	if input.Excerpt != nil {
		blog.Excerpt = *input.Excerpt
	}
	if input.IsPublished != nil {
		wasPublished := blog.IsPublished
		blog.IsPublished = *input.IsPublished
		if !wasPublished && blog.IsPublished && blog.PublishedAt == nil {
			now := time.Now()
			blog.PublishedAt = &now
		}
	}

	blog.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, blog); err != nil {
		return nil, err
	}

	return blog, nil
}

func (s *BlogService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *BlogService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Blog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BlogService) GetBySlug(ctx context.Context, slug string) (*domain.Blog, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *BlogService) List(ctx context.Context, filter domain.BlogFilter) ([]*domain.Blog, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.repo.List(ctx, filter)
}

func generateSlug(title string) string {
	// Simple slug generation
	slug := strings.ToLower(title)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// sanitizeContent removes problematic class attributes from HTML content
// Specifically handles the language-null issue from the editor
func sanitizeContent(content string) string {
	// Remove class="language-null" from code blocks
	content = regexp.MustCompile(`class="language-null"`).ReplaceAllString(content, `class=""`)
	// Also handle language="null" attribute if present
	content = regexp.MustCompile(`language="null"`).ReplaceAllString(content, ``)
	// Clean up empty class attributes
	content = regexp.MustCompile(`class=""\s*`).ReplaceAllString(content, ``)
	return content
}
