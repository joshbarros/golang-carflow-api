package auth

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Role represents user permission levels
type Role string

const (
	// RoleAdmin is for system administrators
	RoleAdmin Role = "admin"
	// RoleUser is for standard users
	RoleUser Role = "user"
)

// User represents a registered user in the system
type User struct {
	ID           string     `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never expose password hash in JSON
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Role         Role       `json:"role" db:"role"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserRegistration represents the data required to register a new user
type UserRegistration struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserLogin represents the data required for user login
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned on successful login
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

// UserProfile contains user information for profile view/edit
type UserProfile struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// SetPassword securely hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	// Hash the password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ValidateRegistration validates the user registration data
func (r *UserRegistration) ValidateRegistration() error {
	// Check email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// Check password strength
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Check required fields
	if r.FirstName == "" {
		return errors.New("first name is required")
	}

	if r.LastName == "" {
		return errors.New("last name is required")
	}

	return nil
}

// ValidateLogin validates the login data
func (l *UserLogin) ValidateLogin() error {
	if l.Email == "" {
		return errors.New("email is required")
	}

	if l.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
