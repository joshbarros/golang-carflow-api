package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) *PostgresRepository {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "carflow",
		Password: "carflow_secret",
		DBName:   "carflow",
	}

	repo, err := NewPostgresRepository(config)
	require.NoError(t, err)
	require.NotNil(t, repo)

	return repo
}

func TestPostgresRepository_CreateUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	user := User{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}

	// Set password
	err := user.SetPassword("password123")
	require.NoError(t, err)

	// Create user
	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)
	assert.NotEmpty(t, createdUser.ID)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.FirstName, createdUser.FirstName)
	assert.Equal(t, user.LastName, createdUser.LastName)
	assert.Equal(t, user.Role, createdUser.Role)
	assert.Equal(t, user.TenantID, createdUser.TenantID)
	assert.True(t, createdUser.Active)
	assert.NotZero(t, createdUser.CreatedAt)
	assert.NotZero(t, createdUser.UpdatedAt)

	// Try to create user with same email
	_, err = repo.CreateUser(user)
	assert.Error(t, err)
}

func TestPostgresRepository_GetUserByID(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	// Create a test user
	user := User{
		Email:     "get-by-id@example.com",
		FirstName: "Get",
		LastName:  "ByID",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}
	err := user.SetPassword("password123")
	require.NoError(t, err)

	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)

	// Get user by ID
	foundUser, err := repo.GetUserByID(createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.Email, foundUser.Email)

	// Try to get non-existent user
	_, err = repo.GetUserByID(uuid.New().String())
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestPostgresRepository_GetUserByEmail(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	// Create a test user
	user := User{
		Email:     "get-by-email@example.com",
		FirstName: "Get",
		LastName:  "ByEmail",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}
	err := user.SetPassword("password123")
	require.NoError(t, err)

	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)

	// Get user by email
	foundUser, err := repo.GetUserByEmail(createdUser.Email)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.Email, foundUser.Email)

	// Try to get non-existent user
	_, err = repo.GetUserByEmail("nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestPostgresRepository_UpdateUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	// Create a test user
	user := User{
		Email:     "update@example.com",
		FirstName: "Update",
		LastName:  "User",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}
	err := user.SetPassword("password123")
	require.NoError(t, err)

	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)

	// Update user
	createdUser.FirstName = "Updated"
	createdUser.LastName = "Name"
	updatedUser, err := repo.UpdateUser(createdUser)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updatedUser.FirstName)
	assert.Equal(t, "Name", updatedUser.LastName)
	assert.True(t, updatedUser.UpdatedAt.After(createdUser.UpdatedAt))
}

func TestPostgresRepository_UpdateLastLogin(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	// Create a test user
	user := User{
		Email:     "last-login@example.com",
		FirstName: "Last",
		LastName:  "Login",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}
	err := user.SetPassword("password123")
	require.NoError(t, err)

	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)

	// Update last login
	time.Sleep(time.Millisecond) // Ensure time difference
	err = repo.UpdateLastLogin(createdUser.ID)
	require.NoError(t, err)

	// Verify last login was updated
	updatedUser, err := repo.GetUserByID(createdUser.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedUser.LastLoginAt)
	assert.True(t, updatedUser.LastLoginAt.After(createdUser.CreatedAt))
}

func TestPostgresRepository_DeleteUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	// Create a test user
	user := User{
		Email:     "delete@example.com",
		FirstName: "Delete",
		LastName:  "User",
		Role:      RoleUser,
		TenantID:  "test-tenant",
		Active:    true,
	}
	err := user.SetPassword("password123")
	require.NoError(t, err)

	createdUser, err := repo.CreateUser(user)
	require.NoError(t, err)

	// Delete user
	err = repo.DeleteUser(createdUser.ID)
	require.NoError(t, err)

	// Verify user was deleted
	_, err = repo.GetUserByID(createdUser.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestPostgresRepository_ListUsersByTenant(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	tenantID := "test-tenant-" + uuid.New().String()

	// Create test users
	for i := 0; i < 3; i++ {
		user := User{
			Email:     fmt.Sprintf("tenant-user-%d@example.com", i),
			FirstName: "Tenant",
			LastName:  fmt.Sprintf("User %d", i),
			Role:      RoleUser,
			TenantID:  tenantID,
			Active:    true,
		}
		err := user.SetPassword("password123")
		require.NoError(t, err)

		_, err = repo.CreateUser(user)
		require.NoError(t, err)
	}

	// Create a user in a different tenant
	otherUser := User{
		Email:     "other-tenant@example.com",
		FirstName: "Other",
		LastName:  "Tenant",
		Role:      RoleUser,
		TenantID:  "other-tenant",
		Active:    true,
	}
	err := otherUser.SetPassword("password123")
	require.NoError(t, err)

	_, err = repo.CreateUser(otherUser)
	require.NoError(t, err)

	// List users by tenant
	users, err := repo.ListUsersByTenant(tenantID)
	require.NoError(t, err)
	assert.Len(t, users, 3)
	for _, user := range users {
		assert.Equal(t, tenantID, user.TenantID)
	}
}
