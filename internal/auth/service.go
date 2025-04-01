package auth

import (
	"time"

	"github.com/google/uuid"
)

// Service provides auth-related operations
type Service struct {
	repo         Repository
	tokenService *TokenService
}

// NewService creates a new auth service
func NewService(repo Repository) *Service {
	return &Service{
		repo:         repo,
		tokenService: NewTokenService(),
	}
}

// Register creates a new user account
func (s *Service) Register(reg UserRegistration, tenantID string) (User, error) {
	// Validate registration data
	if err := reg.ValidateRegistration(); err != nil {
		return User{}, err
	}

	// Create user object
	user := User{
		ID:        uuid.New().String(),
		Email:     reg.Email,
		FirstName: reg.FirstName,
		LastName:  reg.LastName,
		Role:      RoleUser, // Default role
		TenantID:  tenantID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password (will be hashed)
	if err := user.SetPassword(reg.Password); err != nil {
		return User{}, err
	}

	// Save user to repository
	return s.repo.CreateUser(user)
}

// Login authenticates a user and returns JWT tokens
func (s *Service) Login(login UserLogin) (LoginResponse, error) {
	// Validate login data
	if err := login.ValidateLogin(); err != nil {
		return LoginResponse{}, err
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(login.Email)
	if err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.Active {
		return LoginResponse{}, ErrInvalidCredentials
	}

	// Verify password
	if !user.CheckPassword(login.Password) {
		return LoginResponse{}, ErrInvalidCredentials
	}

	// Generate access token
	accessToken, expiresAt, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	// Generate refresh token
	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	// Update last login timestamp
	if err := s.repo.UpdateLastLogin(user.ID); err != nil {
		// Log the error but continue (non-critical)
		// log.Printf("Failed to update last login: %v", err)
	}

	// Return login response
	return LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *Service) RefreshToken(refreshToken string) (string, time.Time, error) {
	return s.tokenService.RefreshAccessToken(refreshToken)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(id string) (User, error) {
	return s.repo.GetUserByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (User, error) {
	return s.repo.GetUserByEmail(email)
}

// UpdateUserProfile updates a user's profile information
func (s *Service) UpdateUserProfile(id string, profile UserProfile) (User, error) {
	// Get current user
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return User{}, err
	}

	// Update fields
	user.FirstName = profile.FirstName
	user.LastName = profile.LastName

	// Only update email if it changed
	if profile.Email != user.Email {
		user.Email = profile.Email
	}

	// Save updated user
	return s.repo.UpdateUser(user)
}

// ChangePassword updates a user's password
func (s *Service) ChangePassword(id, currentPassword, newPassword string) error {
	// Get current user
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return err
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return ErrInvalidCredentials
	}

	// Set new password
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	// Save updated user
	_, err = s.repo.UpdateUser(user)
	return err
}

// DeactivateUser deactivates a user account
func (s *Service) DeactivateUser(id string) error {
	// Get current user
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return err
	}

	// Deactivate account
	user.Active = false
	user.UpdatedAt = time.Now()

	// Save updated user
	_, err = s.repo.UpdateUser(user)
	return err
}

// ListUsersByTenant retrieves all users for a specific tenant
func (s *Service) ListUsersByTenant(tenantID string) ([]User, error) {
	return s.repo.ListUsersByTenant(tenantID)
}

// ValidateToken validates an access token and returns the claims
func (s *Service) ValidateToken(token string) (*TokenClaims, error) {
	return s.tokenService.ValidateAccessToken(token)
}
