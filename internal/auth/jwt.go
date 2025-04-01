package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType defines the type of token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"
	// RefreshToken is used to get new access tokens
	RefreshToken TokenType = "refresh"
)

// TokenClaims extends standard JWT claims with custom fields
type TokenClaims struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     Role      `json:"role"`
	TenantID string    `json:"tenant_id"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenService handles JWT token operations
type TokenService struct {
	accessSecret  string
	refreshSecret string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewTokenService creates a new token service
func NewTokenService() *TokenService {
	// Default values
	accessExpiry := 15 * time.Minute    // 15 minutes by default
	refreshExpiry := 7 * 24 * time.Hour // 7 days by default

	// Get from environment if available
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	// Use defaults for local development if not set
	if accessSecret == "" {
		accessSecret = "access_secret_for_development_only"
	}
	if refreshSecret == "" {
		refreshSecret = "refresh_secret_for_development_only"
	}

	return &TokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates a new access token for a user
func (s *TokenService) GenerateAccessToken(user User) (string, time.Time, error) {
	return s.generateToken(user, AccessToken, s.accessSecret, s.accessExpiry)
}

// GenerateRefreshToken creates a new refresh token for a user
func (s *TokenService) GenerateRefreshToken(user User) (string, time.Time, error) {
	return s.generateToken(user, RefreshToken, s.refreshSecret, s.refreshExpiry)
}

// generateToken is a helper function to create tokens
func (s *TokenService) generateToken(user User, tokenType TokenType, secret string, expiry time.Duration) (string, time.Time, error) {
	expirationTime := time.Now().Add(expiry)

	claims := TokenClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		TenantID: user.TenantID,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "carflow-api",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *TokenService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return s.validateToken(tokenString, s.accessSecret, AccessToken)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *TokenService) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	return s.validateToken(tokenString, s.refreshSecret, RefreshToken)
}

// validateToken is a helper function to validate tokens
func (s *TokenService) validateToken(tokenString, secret string, tokenType TokenType) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// Verify token type matches expected type
		if claims.Type != tokenType {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken creates a new access token from a valid refresh token
func (s *TokenService) RefreshAccessToken(refreshToken string) (string, time.Time, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", time.Time{}, err
	}

	// Create a user object from the claims
	user := User{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		TenantID: claims.TenantID,
	}

	// Generate a new access token
	return s.GenerateAccessToken(user)
}
