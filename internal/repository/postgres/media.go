package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

type MediaRepository struct {
	db *DB
}

func NewMediaRepository(db *DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, input domain.CreateMediaInput) (*domain.Media, error) {
	query := `
		INSERT INTO media (user_id, filename, data, content_type, size_bytes, alt_text, storage_provider)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, filename, content_type, size_bytes, alt_text, storage_provider, file_key, public_url, created_at, updated_at
	`

	m := &domain.Media{}
	// Handle nil data for non-DB storage
	var data interface{} = input.Data
	if len(input.Data) == 0 {
		data = nil
	}

	var fileKey, publicURL *string // Temp vars for nullable strings

	err := r.db.Pool.QueryRow(ctx, query,
		input.UserID,
		input.Filename,
		data,
		input.ContentType,
		input.SizeBytes,
		input.AltText,
		input.StorageProvider,
	).Scan(
		&m.ID,
		&m.UserID,
		&m.Filename,
		&m.ContentType,
		&m.SizeBytes,
		&m.AltText,
		&m.StorageProvider,
		&fileKey,   // May be null
		&publicURL, // May be null
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	if fileKey != nil {
		m.FileKey = *fileKey
	}
	if publicURL != nil {
		m.PublicURL = *publicURL
	}

	// Data is not returned by RETURNING (too large), so we set it back on the struct if provided
	m.Data = input.Data

	return m, nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	query := `
		SELECT id, user_id, filename, data, content_type, size_bytes, alt_text, storage_provider, file_key, public_url, created_at, updated_at
		FROM media
		WHERE id = $1
	`

	m := &domain.Media{}
	var fileKey, publicURL *string // Temp vars for nullable strings

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.UserID,
		&m.Filename,
		&m.Data,
		&m.ContentType,
		&m.SizeBytes,
		&m.AltText,
		&m.StorageProvider,
		&fileKey,
		&publicURL,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	if fileKey != nil {
		m.FileKey = *fileKey
	}
	if publicURL != nil {
		m.PublicURL = *publicURL
	}

	return m, nil
}

func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media WHERE id = $1`

	tag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
