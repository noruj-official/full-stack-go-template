// Package validator provides input validation utilities.
package validator

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail checks if the email format is valid.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsNotEmpty checks if a string is not empty after trimming.
func IsNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// MinLength checks if a string has at least the specified length.
func MinLength(s string, min int) bool {
	return len(s) >= min
}

// MaxLength checks if a string does not exceed the specified length.
func MaxLength(s string, max int) bool {
	return len(s) <= max
}
