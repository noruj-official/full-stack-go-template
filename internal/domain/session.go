// Package domain contains the core business entities and rules.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
type Session struct {
	ID             string    `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionDuration is the default session lifetime.
const SessionDuration = 24 * time.Hour * 7 // 7 days

// NewSession creates a new session for a user.
// NewSession creates a new session for a user.
func NewSession(userID uuid.UUID, ip, userAgent string) *Session {
	now := time.Now()
	return &Session{
		ID:             generateSessionID(),
		UserID:         userID,
		IPAddress:      ip,
		UserAgent:      userAgent,
		ExpiresAt:      now.Add(SessionDuration),
		CreatedAt:      now,
		LastActivityAt: now,
	}
}

// generateSessionID creates a cryptographically secure session ID.
func generateSessionID() string {
	return uuid.New().String()
}
