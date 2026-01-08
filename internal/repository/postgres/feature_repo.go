package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

// FeatureRepository implements the repository interface for Feature Flags.
type FeatureRepository struct {
	db *DB
}

// NewFeatureRepository creates a new feature repository.
func NewFeatureRepository(db *DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

// Get retrieves a feature flag by name.
func (r *FeatureRepository) Get(ctx context.Context, name string) (*domain.FeatureFlag, error) {
	query := `
		SELECT name, enabled, description, created_at, updated_at
		FROM feature_flags
		WHERE name = $1
	`

	feature := &domain.FeatureFlag{}
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&feature.Name,
		&feature.Enabled,
		&feature.Description,
		&feature.CreatedAt,
		&feature.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return feature, nil
}

// List retrieves all feature flags.
func (r *FeatureRepository) List(ctx context.Context) ([]*domain.FeatureFlag, error) {
	query := `
		SELECT name, enabled, description, created_at, updated_at
		FROM feature_flags
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []*domain.FeatureFlag
	for rows.Next() {
		feature := &domain.FeatureFlag{}
		if err := rows.Scan(
			&feature.Name,
			&feature.Enabled,
			&feature.Description,
			&feature.CreatedAt,
			&feature.UpdatedAt,
		); err != nil {
			return nil, err
		}
		features = append(features, feature)
	}

	return features, nil
}

// Upsert creates or updates a feature flag.
func (r *FeatureRepository) Upsert(ctx context.Context, feature *domain.FeatureFlag) error {
	feature.UpdatedAt = time.Now()
	if feature.CreatedAt.IsZero() {
		feature.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO feature_flags (name, enabled, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name) DO UPDATE
		SET enabled = EXCLUDED.enabled,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Pool.Exec(ctx, query,
		feature.Name,
		feature.Enabled,
		feature.Description,
		feature.CreatedAt,
		feature.UpdatedAt,
	)

	return err
}

// Delete removes a feature flag.
func (r *FeatureRepository) Delete(ctx context.Context, name string) error {
	query := `DELETE FROM feature_flags WHERE name = $1`
	result, err := r.db.Pool.Exec(ctx, query, name)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
