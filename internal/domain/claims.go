package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

// EmailAuthClaims defines the claims for email authentication tokens.
type EmailAuthClaims struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

// Token purposes
const (
	TokenPurposeEmailAuth = "email_auth"
)
