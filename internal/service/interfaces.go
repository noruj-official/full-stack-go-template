// Package service defines the interfaces for business logic operations.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
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
}

// EmailService defines the interface for email operations.
type EmailService interface {
	// SendVerificationEmail sends a verification email to the user.
	SendVerificationEmail(ctx context.Context, emailAddr, name, token string) error
}
