package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/repository/postgres"
)

type MediaService struct {
	repo *postgres.MediaRepository
}

func NewMediaService(repo *postgres.MediaRepository) *MediaService {
	return &MediaService{repo: repo}
}

func (s *MediaService) Upload(ctx context.Context, input domain.CreateMediaInput) (*domain.Media, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validate input: %w", err)
	}

	return s.repo.Create(ctx, input)
}

func (s *MediaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ProcessImageUpload reads the file into a byte slice and detects its content type
func ProcessImageUpload(file multipart.File, header *multipart.FileHeader) ([]byte, string, int64, error) {
	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(data)

	return data, contentType, header.Size, nil
}
