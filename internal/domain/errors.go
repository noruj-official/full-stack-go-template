// Package domain contains the core business entities and rules.
package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors for common domain error cases.
var (
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("access forbidden")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionExpired     = errors.New("session has expired")
	ErrInvalidToken       = errors.New("invalid or malformed token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrEmailNotVerified   = errors.New("email not verified")
)

// ErrValidation represents a validation error for a specific field.
type ErrValidation struct {
	Field   string
	Message string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var validationErr ErrValidation
	return errors.As(err, &validationErr)
}

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflictError checks if an error is a conflict error.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsUnauthorizedError checks if an error is an unauthorized error.
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if an error is a forbidden error.
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInvalidCredentialsError checks if an error is an invalid credentials error.
func IsInvalidCredentialsError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}
