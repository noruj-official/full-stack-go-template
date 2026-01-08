package service

import (
	"context"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/repository/postgres"
)

// featureService implements the FeatureService interface.
type featureService struct {
	repo *postgres.FeatureRepository
}

// NewFeatureService creates a new feature service.
func NewFeatureService(repo *postgres.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

// IsEnabled checks if a feature flag is enabled.
func (s *featureService) IsEnabled(ctx context.Context, name string) (bool, error) {
	feature, err := s.repo.Get(ctx, name)
	if err != nil {
		if err == domain.ErrNotFound {
			return false, nil // Default to false if not found
		}
		return false, err
	}
	return feature.Enabled, nil
}

// GetAll retrieves all feature flags.
func (s *featureService) GetAll(ctx context.Context) ([]*domain.FeatureFlag, error) {
	return s.repo.List(ctx)
}

// Toggle enables or disables a feature flag.
func (s *featureService) Toggle(ctx context.Context, name string, enabled bool) error {
	feature, err := s.repo.Get(ctx, name)
	if err != nil {
		if err == domain.ErrNotFound {
			// If not found, create it
			feature = domain.NewFeatureFlag(name, enabled, "")
		} else {
			return err
		}
	} else {
		// Update existing
		feature.Enabled = enabled
	}

	return s.repo.Upsert(ctx, feature)
}

// Upsert creates or updates a feature flag with full details.
func (s *featureService) Upsert(ctx context.Context, name, description string, enabled bool) error {
	feature := domain.NewFeatureFlag(name, enabled, description)
	return s.repo.Upsert(ctx, feature)
}

// SyncFeatures ensures that the specified features exist in the database.
// If a feature does not exist, it is created with the specified default enabled state.
func (s *featureService) SyncFeatures(ctx context.Context, features map[string]domain.FeatureConfig) error {
	for name, config := range features {
		_, err := s.repo.Get(ctx, name)
		if err != nil {
			if err == domain.ErrNotFound {
				// Feature missing, create it
				if err := s.Upsert(ctx, name, config.Description, config.DefaultEnabled); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}
