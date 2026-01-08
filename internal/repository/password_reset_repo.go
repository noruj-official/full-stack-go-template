package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

// PasswordResetRepository defines the interface for password reset token storage.
type PasswordResetRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	GetByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
