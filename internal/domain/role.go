// Package domain contains the core business entities and rules.
package domain

// Role represents the user's permission level in the system.
type Role string

const (
	// RoleUser is a regular user with basic permissions.
	RoleUser Role = "user"

	// RoleAdmin has elevated permissions for user management.
	RoleAdmin Role = "admin"

	// RoleSuperAdmin has full system access including admin management.
	RoleSuperAdmin Role = "super_admin"
)

// IsValid checks if the role is a valid role type.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin, RoleSuperAdmin:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// HasPermission checks if this role has at least the permission level of the required role.
// Permission hierarchy: super_admin > admin > user
func (r Role) HasPermission(required Role) bool {
	roleLevel := map[Role]int{
		RoleUser:       1,
		RoleAdmin:      2,
		RoleSuperAdmin: 3,
	}

	return roleLevel[r] >= roleLevel[required]
}

// CanManageRole checks if this role can manage (create/edit/delete) the target role.
func (r Role) CanManageRole(target Role) bool {
	// Super admin can manage anyone
	if r == RoleSuperAdmin {
		return true
	}

	// Admin can only manage users, not other admins
	if r == RoleAdmin && target == RoleUser {
		return true
	}

	return false
}

// AllRoles returns all valid roles.
func AllRoles() []Role {
	return []Role{RoleUser, RoleAdmin, RoleSuperAdmin}
}
