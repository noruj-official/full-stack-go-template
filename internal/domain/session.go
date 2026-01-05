// Package domain contains the core business entities and rules.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
type Session struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionDuration is the default session lifetime.
const SessionDuration = 24 * time.Hour * 7 // 7 days

// NewSession creates a new session for a user.
func NewSession(userID uuid.UUID) *Session {
	return &Session{
		ID:        generateSessionID(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(SessionDuration),
		CreatedAt: time.Now(),
	}
}

// generateSessionID creates a cryptographically secure session ID.
func generateSessionID() string {
	return uuid.New().String() + uuid.New().String()
}
