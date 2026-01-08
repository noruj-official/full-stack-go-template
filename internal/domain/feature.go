package domain

import "time"

// FeatureFlag represents a toggleable feature in the system.
type FeatureFlag struct {
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	FeatureThemeManagement = "theme_management"
	FeatureEmailAuth       = "email_auth"
)

// FeatureConfig represents the initial configuration for a feature flag.
type FeatureConfig struct {
	Description    string
	DefaultEnabled bool
}

// NewFeatureFlag creates a new FeatureFlag.
func NewFeatureFlag(name string, enabled bool, description string) *FeatureFlag {
	now := time.Now()
	return &FeatureFlag{
		Name:        name,
		Enabled:     enabled,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
