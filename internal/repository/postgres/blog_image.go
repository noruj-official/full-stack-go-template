package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

type BlogImageRepository struct {
	db *DB
}

func NewBlogImageRepository(db *DB) *BlogImageRepository {
	return &BlogImageRepository{db: db}
}

func (r *BlogImageRepository) Create(ctx context.Context, img *domain.BlogImage) error {
	query := `
		INSERT INTO blog_images (id, blog_id, media_id, alt_text, caption, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		img.ID, img.BlogID, img.MediaID,
		img.AltText, img.Caption, img.Position, img.CreatedAt, img.UpdatedAt,
	)
	return err
}

func (r *BlogImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BlogImage, error) {
	query := `
		SELECT id, blog_id, media_id, alt_text, caption, position, created_at, updated_at
		FROM blog_images
		WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var img domain.BlogImage
	err := row.Scan(
		&img.ID, &img.BlogID, &img.MediaID,
		&img.AltText, &img.Caption, &img.Position, &img.CreatedAt, &img.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &img, nil
}

// GetByIDWithoutData retrieves image metadata (same as GetByID now as data is separate)
func (r *BlogImageRepository) GetByIDWithoutData(ctx context.Context, id uuid.UUID) (*domain.BlogImage, error) {
	return r.GetByID(ctx, id)
}

func (r *BlogImageRepository) ListByBlogID(ctx context.Context, blogID uuid.UUID) ([]*domain.BlogImage, error) {
	query := `
		SELECT id, blog_id, media_id, alt_text, caption, position, created_at, updated_at
		FROM blog_images
		WHERE blog_id = $1
		ORDER BY position ASC, created_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*domain.BlogImage
	for rows.Next() {
		var img domain.BlogImage
		err := rows.Scan(
			&img.ID, &img.BlogID, &img.MediaID,
			&img.AltText, &img.Caption, &img.Position, &img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, &img)
	}

	return images, nil
}

func (r *BlogImageRepository) Update(ctx context.Context, img *domain.BlogImage) error {
	query := `
		UPDATE blog_images
		SET alt_text = $1, caption = $2, position = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query,
		img.AltText, img.Caption, img.Position, time.Now(), img.ID,
	)
	return err
}

func (r *BlogImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM blog_images WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *BlogImageRepository) GetCoverImage(ctx context.Context, blogID uuid.UUID) ([]byte, string, error) {
	query := `
		SELECT m.data, m.content_type 
		FROM blogs b
		JOIN media m ON b.cover_media_id = m.id
		WHERE b.id = $1
	`

	var imageData []byte
	var imageType string
	err := r.db.Pool.QueryRow(ctx, query, blogID).Scan(&imageData, &imageType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}

	return imageData, imageType, nil
}

func (r *BlogImageRepository) GetCoverImageBySlug(ctx context.Context, slug string) ([]byte, string, error) {
	query := `
		SELECT m.data, m.content_type 
		FROM blogs b
		JOIN media m ON b.cover_media_id = m.id
		WHERE b.slug = $1
	`

	var imageData []byte
	var imageType string
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(&imageData, &imageType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}

	return imageData, imageType, nil
}
