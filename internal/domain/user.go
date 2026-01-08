// Package domain contains the core business entities and rules.
// These types are independent of any external frameworks or databases.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

// User represents a user entity in the system.
type User struct {
	ID                         uuid.UUID  `json:"id"`
	Email                      string     `json:"email"`
	Name                       string     `json:"name"`
	PasswordHash               string     `json:"-"` // Never expose in JSON
	Role                       Role       `json:"role"`
	ProfileImage               []byte     `json:"-"`                  // Binary image data (never expose in JSON)
	ProfileImageType           string     `json:"profile_image_type"` // MIME type (e.g., "image/jpeg")
	ProfileImageSize           int        `json:"profile_image_size"` // Size in bytes
	EmailVerified              bool       `json:"email_verified"`
	VerificationToken          *string    `json:"-"`
	VerificationTokenExpiresAt *time.Time `json:"-"`
	Status                     UserStatus `json:"status"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
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
		Status:       UserStatusActive,
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

// UpdateProfileImageInput represents the input for updating a user's profile image.
type UpdateProfileImageInput struct {
	ImageData   []byte
	ContentType string
	Size        int
}

// Validate checks if the profile image input is valid.
func (i *UpdateProfileImageInput) Validate() error {
	const maxSize = 2 * 1024 * 1024 // 2MB in bytes

	if len(i.ImageData) == 0 {
		return ErrValidation{Field: "image", Message: "image data is required"}
	}

	if i.Size > maxSize {
		return ErrValidation{Field: "image", Message: "image size must be less than 2MB"}
	}

	// Validate content type
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validTypes[i.ContentType] {
		return ErrValidation{Field: "image", Message: "invalid image type. Allowed: JPEG, PNG, GIF, WebP"}
	}

	return nil
}
