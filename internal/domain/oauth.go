package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthProviderType represents the type of OAuth provider.
type OAuthProviderType string

const (
	OAuthProviderGoogle   OAuthProviderType = "google"
	OAuthProviderGitHub   OAuthProviderType = "github"
	OAuthProviderLinkedIn OAuthProviderType = "linkedin"
)

// OAuthProvider represents an OAuth provider configuration.
type OAuthProvider struct {
	Provider     OAuthProviderType `json:"provider"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"` // Should be protected
	Enabled      bool              `json:"enabled"`
	Scopes       []string          `json:"scopes"`
	AuthURL      string            `json:"auth_url"`
	TokenURL     string            `json:"token_url"`
	UserInfoURL  string            `json:"user_info_url"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// UserOAuth represents a link between a user and an OAuth provider.
type UserOAuth struct {
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	Provider       OAuthProviderType `json:"provider"`
	ProviderUserID string            `json:"provider_user_id"`
	AccessToken    string            `json:"-"`
	RefreshToken   string            `json:"-"`
	ExpiresAt      *time.Time        `json:"expires_at"`
	CreatedAt      time.Time         `json:"created_at"`
}

// OAuthCallbackParams represents the parameters returned from an OAuth callback.
type OAuthCallbackParams struct {
	Code  string
	State string
}

// OAuthUserInfo represents the user info retrieved from the provider.
type OAuthUserInfo struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

// UpdateOAuthProviderInput represents the input for updating an OAuth provider.
type UpdateOAuthProviderInput struct {
	ClientID     *string   `json:"client_id"`
	ClientSecret *string   `json:"client_secret"`
	Enabled      *bool     `json:"enabled"`
	Scopes       *[]string `json:"scopes"`
}
