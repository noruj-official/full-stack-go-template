// Package domain contains the core business entities and rules.
// These types are independent of any external frameworks or databases.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewUser creates a new User with a generated UUID and timestamps.
func NewUser(email, name, passwordHash string, role Role) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate checks if the user data is valid.
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrValidation{Field: "email", Message: "email is required"}
	}
	if u.Name == "" {
		return ErrValidation{Field: "name", Message: "name is required"}
	}
	if !u.Role.IsValid() {
		return ErrValidation{Field: "role", Message: "invalid role"}
	}
	return nil
}

// IsAdmin checks if user has admin or super_admin role.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsSuperAdmin checks if user has super_admin role.
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// HasPermission checks if user has at least the required permission level.
func (u *User) HasPermission(required Role) bool {
	return u.Role.HasPermission(required)
}

// RegisterInput represents the input for user registration.
type RegisterInput struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// Validate checks if the registration input is valid.
func (i *RegisterInput) Validate() error {
	if i.Email == "" {
		return ErrValidation{Field: "email", Message: "email is required"}
	}
	if i.Name == "" {
		return ErrValidation{Field: "name", Message: "name is required"}
	}
	if i.Password == "" {
		return ErrValidation{Field: "password", Message: "password is required"}
	}
	if len(i.Password) < 8 {
		return ErrValidation{Field: "password", Message: "password must be at least 8 characters"}
	}
	if i.Password != i.ConfirmPassword {
		return ErrValidation{Field: "confirm_password", Message: "passwords do not match"}
	}
	return nil
}

// LoginInput represents the input for user login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks if the login input is valid.
func (i *LoginInput) Validate() error {
	if i.Email == "" {
		return ErrValidation{Field: "email", Message: "email is required"}
	}
	if i.Password == "" {
		return ErrValidation{Field: "password", Message: "password is required"}
	}
	return nil
}

// CreateUserInput represents the input for creating a new user (admin action).
type CreateUserInput struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

// Validate checks if the create user input is valid.
func (i *CreateUserInput) Validate() error {
	if i.Email == "" {
		return ErrValidation{Field: "email", Message: "email is required"}
	}
	if i.Name == "" {
		return ErrValidation{Field: "name", Message: "name is required"}
	}
	if i.Password == "" {
		return ErrValidation{Field: "password", Message: "password is required"}
	}
	if len(i.Password) < 8 {
		return ErrValidation{Field: "password", Message: "password must be at least 8 characters"}
	}
	if !i.Role.IsValid() {
		return ErrValidation{Field: "role", Message: "invalid role"}
	}
	return nil
}

// UpdateUserInput represents the input for updating a user.
type UpdateUserInput struct {
	Email *string `json:"email,omitempty"`
	Name  *string `json:"name,omitempty"`
	Role  *Role   `json:"role,omitempty"`
}
