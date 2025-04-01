package auth

import "errors"

// Common errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email is already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Repository defines the interface for user data access
type Repository interface {
	CreateUser(user User) (User, error)
	GetUserByID(id string) (User, error)
	GetUserByEmail(email string) (User, error)
	UpdateUser(user User) (User, error)
	UpdateLastLogin(id string) error
	ListUsersByTenant(tenantID string) ([]User, error)
}
