package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a token used for resetting a user's password.
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the token has expired.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NewPasswordResetToken creates a new password reset token.
func NewPasswordResetToken(userID uuid.UUID, hash string, duration time.Duration) *PasswordResetToken {
	now := time.Now()
	return &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: now.Add(duration),
		CreatedAt: now,
	}
}
