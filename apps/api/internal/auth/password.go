package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 10
)

var (
	// ErrPasswordTooShort is returned when the password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrInvalidPassword is returned when the password is invalid
	ErrInvalidPassword = errors.New("invalid password")
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to check password: %w", err)
	}
	return nil
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	// Add more validation rules as needed:
	// - Must contain uppercase
	// - Must contain lowercase
	// - Must contain number
	// - Must contain special character

	return nil
}
