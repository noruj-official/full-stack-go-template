// Package service defines the interfaces for business logic operations.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

// UserService defines the interface for user business operations.
type UserService interface {
	// CreateUser creates a new user with validation.
	CreateUser(ctx context.Context, input *domain.CreateUserInput) (*domain.User, error)

	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// ListUsers retrieves all users with pagination.
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)

	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, id uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error)

	// UpdateStatus updates the status of a user.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error

	// UpdatePassword updates the user's password.
	UpdatePassword(ctx context.Context, id uuid.UUID, input *domain.UpdatePasswordInput) error

	// DeleteUser removes a user.
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// AuthService defines the interface for authentication operations.
type AuthService interface {
	// Register creates a new user account.
	Register(ctx context.Context, input *domain.RegisterInput, ip, userAgent string) (*domain.User, error)

	// Login authenticates a user and creates a session.
	Login(ctx context.Context, input *domain.LoginInput, ip, userAgent string) (*domain.User, *domain.Session, error)

	// Logout destroys a user session.
	Logout(ctx context.Context, sessionID string) error

	// ValidateSession checks if a session is valid and returns the user.
	ValidateSession(ctx context.Context, sessionID string) (*domain.User, error)

	// GetCurrentUser retrieves the authenticated user from session.
	GetCurrentUser(ctx context.Context, sessionID string) (*domain.User, error)

	// VerifyEmail verifies a user's email address using a token.
	VerifyEmail(ctx context.Context, token string) error

	// RequestPasswordReset initiates the password reset flow.
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword resets the user's password using the token.
	ResetPassword(ctx context.Context, token, newPassword string) error

	// SignOutAllDevices invalidates all sessions for a user.
	SignOutAllDevices(ctx context.Context, userID uuid.UUID) error

	// GetOAuthLoginURL generates a login URL for the specified provider.
	GetOAuthLoginURL(ctx context.Context, provider domain.OAuthProviderType, state string) (string, error)

	// LoginWithOAuth handles the OAuth callback and logs in the user.
	LoginWithOAuth(ctx context.Context, provider domain.OAuthProviderType, code, state string, ip, userAgent string) (*domain.User, *domain.Session, error)

	// ListEnabledProviders returns a map of enabled providers.
	ListEnabledProviders(ctx context.Context) (map[string]bool, error)

	// GenerateEmailAuthToken generates a signed JWT for email authentication.
	GenerateEmailAuthToken(email string, purpose string) (string, error)

	// VerifyEmailAuthToken verifies a signed JWT and returns the claims.
	VerifyEmailAuthToken(token string) (string, string, error)

	// LoginWithEmailToken authenticates a user using a magic link token.
	LoginWithEmailToken(ctx context.Context, token string, ip, userAgent string) (*domain.User, *domain.Session, error)

	// SendEmailAuthLink sends a magic link email to the user.
	SendEmailAuthLink(ctx context.Context, emailAddr, token string) error
}

// EmailService defines the interface for email operations.
type EmailService interface {
	// SendVerificationEmail sends a verification email to the user.
	SendVerificationEmail(ctx context.Context, emailAddr, name, token string) error

	// SendPasswordResetEmail sends a password reset email to the user.
	SendPasswordResetEmail(ctx context.Context, emailAddr, name, token string) error

	// SendEmailAuthLink sends a magic link email to the user.
	SendEmailAuthLink(ctx context.Context, emailAddr, token string) error
}

// FeatureService defines the interface for feature flag operations.
type FeatureService interface {
	// IsEnabled checks if a feature flag is enabled.
	IsEnabled(ctx context.Context, name string) (bool, error)

	// GetAll retrieves all feature flags.
	GetAll(ctx context.Context) ([]*domain.FeatureFlag, error)

	// Toggle enables or disables a feature flag.
	Toggle(ctx context.Context, name string, enabled bool) error

	// SyncFeatures synchronizes feature flags from a map.
	SyncFeatures(ctx context.Context, features map[string]domain.FeatureConfig) error

	// Upsert creates or updates a feature flag with full details.
	Upsert(ctx context.Context, name, description string, enabled bool) error
}
