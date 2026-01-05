// Package repository defines the interfaces for data access.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
)

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	// Create inserts a new user into the database.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// List retrieves all users with optional pagination.
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)

	// Update modifies an existing user in the database.
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user from the database.
	Delete(ctx context.Context, id uuid.UUID) error

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)
}

// SessionRepository defines the interface for session data access operations.
type SessionRepository interface {
	// Create inserts a new session into the database.
	Create(ctx context.Context, session *domain.Session) error

	// GetByID retrieves a session by its ID.
	GetByID(ctx context.Context, id string) (*domain.Session, error)

	// Delete removes a session by its ID.
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired removes all expired sessions.
	DeleteExpired(ctx context.Context) error

	// CountActive returns the number of active (non-expired) sessions.
	CountActive(ctx context.Context) (int64, error)
}
