package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

type BlogImageRepository interface {
	Create(ctx context.Context, img *domain.BlogImage) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogImage, error)
	GetByIDWithoutData(ctx context.Context, id uuid.UUID) (*domain.BlogImage, error)
	ListByBlogID(ctx context.Context, blogID uuid.UUID) ([]*domain.BlogImage, error)
	Update(ctx context.Context, img *domain.BlogImage) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetCoverImage(ctx context.Context, blogID uuid.UUID) ([]byte, string, error)
	GetCoverImageBySlug(ctx context.Context, slug string) ([]byte, string, error)
}

type BlogImageService struct {
	repo         BlogImageRepository
	mediaService *MediaService
}

func NewBlogImageService(repo BlogImageRepository, mediaService *MediaService) *BlogImageService {
	return &BlogImageService{repo: repo, mediaService: mediaService}
}

func (s *BlogImageService) Upload(ctx context.Context, input domain.CreateBlogImageInput) (*domain.BlogImage, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Upload to MediaService
	mediaInput := domain.CreateMediaInput{
		UserID:      nil,                 // TODO: associate with blog author if possible, or leave nil for now (system/blog owned)
		Filename:    "gallery_image.jpg", // TODO: capture original filename in input?
		Data:        input.ImageData,
		ContentType: input.ContentType,
		SizeBytes:   len(input.ImageData),
		AltText:     input.AltText,
	}

	media, err := s.mediaService.Upload(ctx, mediaInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}

	now := time.Now()
	img := &domain.BlogImage{
		ID:        uuid.New(),
		BlogID:    input.BlogID,
		MediaID:   media.ID,
		Media:     media,
		AltText:   input.AltText,
		Caption:   input.Caption,
		Position:  input.Position,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, img); err != nil {
		return nil, err
	}

	return img, nil
}

func (s *BlogImageService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogImage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BlogImageService) GetImageData(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	img, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if img == nil {
		return nil, "", fmt.Errorf("image not found")
	}

	media, err := s.mediaService.GetByID(ctx, img.MediaID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get media: %w", err)
	}

	return media.Data, media.ContentType, nil
}

func (s *BlogImageService) List(ctx context.Context, blogID uuid.UUID) ([]*domain.BlogImage, error) {
	return s.repo.ListByBlogID(ctx, blogID)
}

func (s *BlogImageService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateBlogImageInput) (*domain.BlogImage, error) {
	img, err := s.repo.GetByIDWithoutData(ctx, id)
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, fmt.Errorf("image not found")
	}

	if input.AltText != nil {
		img.AltText = *input.AltText
	}
	if input.Caption != nil {
		img.Caption = *input.Caption
	}
	if input.Position != nil {
		img.Position = *input.Position
	}

	img.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, img); err != nil {
		return nil, err
	}

	return img, nil
}

func (s *BlogImageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *BlogImageService) GetCoverImage(ctx context.Context, blogID uuid.UUID) ([]byte, string, error) {
	return s.repo.GetCoverImage(ctx, blogID)
}

func (s *BlogImageService) GetCoverImageBySlug(ctx context.Context, slug string) ([]byte, string, error) {
	return s.repo.GetCoverImageBySlug(ctx, slug)
}

// GetMediaByID gets media directly by media ID (for serving cover images)
func (s *BlogImageService) GetMediaByID(ctx context.Context, mediaID uuid.UUID) (*domain.Media, error) {
	return s.mediaService.GetByID(ctx, mediaID)
}

// ValidateImageType checks if the image type is allowed
func ValidateImageType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return allowedTypes[contentType]
}

// ValidateImageSize checks if the image size is within limits (5MB)
func ValidateImageSize(size int) bool {
	maxSize := 5 * 1024 * 1024 // 5MB
	return size <= maxSize
}

// ProcessImageUpload processes a multipart file upload
func ProcessImageUpload(file multipart.File, header *multipart.FileHeader) ([]byte, string, int, error) {
	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to infer from filename
		if strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") || strings.HasSuffix(strings.ToLower(header.Filename), ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(strings.ToLower(header.Filename), ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(strings.ToLower(header.Filename), ".gif") {
			contentType = "image/gif"
		} else if strings.HasSuffix(strings.ToLower(header.Filename), ".webp") {
			contentType = "image/webp"
		}
	}

	if !ValidateImageType(contentType) {
		return nil, "", 0, fmt.Errorf("invalid image type: %s. Allowed types: JPEG, PNG, GIF, WebP", contentType)
	}

	// Validate file size
	size := int(header.Size)
	if !ValidateImageSize(size) {
		return nil, "", 0, fmt.Errorf("image size exceeds 5MB limit")
	}

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to read image data: %w", err)
	}

	return data, contentType, size, nil
}
