package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-carflow-api/internal/tenant"
	"github.com/joshbarros/golang-carflow-api/internal/user"
)

var (
	// ErrEmailAlreadyExists is returned when email is already registered
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotActive is returned when user account is disabled
	ErrUserNotActive = errors.New("user account is not active")
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
)

// Service handles authentication business logic
type Service struct {
	userRepo   user.Repository
	tenantRepo tenant.Repository
}

// NewService creates a new auth service
func NewService(userRepo user.Repository, tenantRepo tenant.Repository) *Service {
	return &Service{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// RegisterRequest contains registration data
type RegisterRequest struct {
	DealershipName string `json:"dealership_name"`
	OwnerEmail     string `json:"owner_email"`
	OwnerPassword  string `json:"owner_password"`
	OwnerFirstName string `json:"owner_first_name"`
	OwnerLastName  string `json:"owner_last_name"`
	OwnerPhone     string `json:"owner_phone,omitempty"`
	CompanyAddress string `json:"company_address,omitempty"`
	Plan           string `json:"plan,omitempty"`
}

// RegisterResponse contains registration response data
type RegisterResponse struct {
	TenantID string      `json:"tenant_id"`
	UserID   string      `json:"user_id"`
	Token    string      `json:"token"`
	User     *user.User  `json:"user"`
	Tenant   *tenant.Tenant `json:"tenant"`
}

// LoginRequest contains login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains login response data
type LoginResponse struct {
	Token  string      `json:"token"`
	User   *user.User  `json:"user"`
	Tenant *tenant.Tenant `json:"tenant"`
}

// Register creates a new tenant and owner user
func (s *Service) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// Validate input
	if err := validateEmail(req.OwnerEmail); err != nil {
		return nil, err
	}
	if err := ValidatePasswordStrength(req.OwnerPassword); err != nil {
		return nil, err
	}
	if req.DealershipName == "" {
		return nil, fmt.Errorf("dealership name is required")
	}
	if req.OwnerFirstName == "" || req.OwnerLastName == "" {
		return nil, fmt.Errorf("owner first and last name are required")
	}

	// Check if email already exists
	existingTenant, err := s.tenantRepo.GetByEmail(req.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingTenant != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Determine subscription plan
	plan := req.Plan
	if plan == "" {
		plan = tenant.PlanStarter
	}

	// Get plan limits
	maxVehicles, maxUsers, maxAPICalls := tenant.PlanLimits(plan)

	// Generate tenant slug from dealership name
	slug := generateSlug(req.DealershipName)

	// Check if slug is unique (add random suffix if not)
	existingSlug, _ := s.tenantRepo.GetBySlug(slug)
	if existingSlug != nil {
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	}

	// Create tenant
	newTenant := &tenant.Tenant{
		ID:                     uuid.New().String(),
		Name:                   req.DealershipName,
		Slug:                   slug,
		SubscriptionStatus:     tenant.StatusTrial,
		SubscriptionPlan:       plan,
		TrialEndsAt:            timePtr(time.Now().AddDate(0, 0, 14)), // 14 days trial
		MaxVehicles:            maxVehicles,
		MaxUsers:               maxUsers,
		MaxAPICallsPerMonth:    maxAPICalls,
		OwnerEmail:             req.OwnerEmail,
		OwnerPhone:             req.OwnerPhone,
		CompanyAddress:         req.CompanyAddress,
	}

	err = s.tenantRepo.Create(newTenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Hash password
	passwordHash, err := HashPassword(req.OwnerPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create owner user
	newUser := &user.User{
		ID:           uuid.New().String(),
		TenantID:     newTenant.ID,
		Email:        req.OwnerEmail,
		PasswordHash: passwordHash,
		FirstName:    req.OwnerFirstName,
		LastName:     req.OwnerLastName,
		Role:         user.RoleOwner,
		IsActive:     true,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := GenerateToken(newUser.ID, newTenant.ID, newUser.Email, newUser.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterResponse{
		TenantID: newTenant.ID,
		UserID:   newUser.ID,
		Token:    token,
		User:     newUser,
		Tenant:   newTenant,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	// Validate input
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Find tenant by email
	foundTenant, err := s.tenantRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}
	if foundTenant == nil {
		return nil, ErrInvalidCredentials
	}

	// Find user by email within tenant
	foundUser, err := s.userRepo.GetByEmail(foundTenant.ID, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !foundUser.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	err = CheckPassword(req.Password, foundUser.PasswordHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login timestamp
	_ = s.userRepo.UpdateLastLogin(foundUser.ID)

	// Generate JWT token
	token, err := GenerateToken(foundUser.ID, foundTenant.ID, foundUser.Email, foundUser.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:  token,
		User:   foundUser,
		Tenant: foundTenant,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID string) (*user.User, error) {
	return s.userRepo.GetByID(userID)
}

// generateSlug generates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

// validateEmail validates email format
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
