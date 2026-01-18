package service

import (
	"context"
	"fmt"

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
