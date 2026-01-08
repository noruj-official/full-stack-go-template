package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

// PasswordResetRepository implements the repository.PasswordResetRepository interface.
type PasswordResetRepository struct {
	db *DB
}

// NewPasswordResetRepository creates a new PostgreSQL password reset repository.
func NewPasswordResetRepository(db *DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create inserts a new password reset token.
func (r *PasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

// GetByHash retrieves a token by its hash.
func (r *PasswordResetRepository) GetByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	token := &domain.PasswordResetToken{}
	err := r.db.Pool.QueryRow(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return token, nil
}

// Delete removes a token by its ID.
func (r *PasswordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// DeleteExpired removes all expired tokens.
func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < NOW()`
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}
