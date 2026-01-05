// Package domain contains the core business entities and rules.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ActivityType represents the type of user activity.
type ActivityType string

const (
	// ActivityLogin represents a user login event.
	ActivityLogin ActivityType = "login"

	// ActivityLogout represents a user logout event.
	ActivityLogout ActivityType = "logout"

	// ActivityProfileUpdate represents a profile update event.
	ActivityProfileUpdate ActivityType = "profile_update"

	// ActivityPasswordChange represents a password change event.
	ActivityPasswordChange ActivityType = "password_change"

	// ActivitySettingsUpdate represents a settings update event.
	ActivitySettingsUpdate ActivityType = "settings_update"
)

// ActivityLog represents a user activity log entry.
type ActivityLog struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	ActivityType ActivityType `json:"activity_type"`
	Description  string       `json:"description"`
	IPAddress    *string      `json:"ip_address,omitempty"`
	UserAgent    *string      `json:"user_agent,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

// AuditAction represents an administrative action type.
type AuditAction string

const (
	// AuditUserCreate represents user creation.
	AuditUserCreate AuditAction = "user.create"

	// AuditUserUpdate represents user update.
	AuditUserUpdate AuditAction = "user.update"

	// AuditUserDelete represents user deletion.
	AuditUserDelete AuditAction = "user.delete"

	// AuditRoleChange represents role change.
	AuditRoleChange AuditAction = "user.role_change"

	// AuditSystemConfig represents system configuration change.
	AuditSystemConfig AuditAction = "system.config_change"
)

// AuditLog represents an audit log entry for administrative actions.
type AuditLog struct {
	ID           uuid.UUID              `json:"id"`
	AdminID      uuid.UUID              `json:"admin_id"`
	AdminName    string                 `json:"admin_name,omitempty"` // Populated via join
	Action       AuditAction            `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}
