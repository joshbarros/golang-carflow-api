package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose password hash in JSON
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Role        string     `json:"role" db:"role"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// HasRole checks if the user has the specified role
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsOwner checks if the user is an owner
func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

// IsAdmin checks if the user is an admin or owner
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleOwner
}
