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
		INSERT INTO media (user_id, filename, data, content_type, size_bytes, alt_text)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, filename, content_type, size_bytes, alt_text, created_at, updated_at
	`

	m := &domain.Media{}
	err := r.db.Pool.QueryRow(ctx, query,
		input.UserID,
		input.Filename,
		input.Data,
		input.ContentType,
		input.SizeBytes,
		input.AltText,
	).Scan(
		&m.ID,
		&m.UserID,
		&m.Filename,
		&m.ContentType,
		&m.SizeBytes,
		&m.AltText,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return m, nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	query := `
		SELECT id, user_id, filename, data, content_type, size_bytes, alt_text, created_at, updated_at
		FROM media
		WHERE id = $1
	`

	m := &domain.Media{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.UserID,
		&m.Filename,
		&m.Data,
		&m.ContentType,
		&m.SizeBytes,
		&m.AltText,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
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
