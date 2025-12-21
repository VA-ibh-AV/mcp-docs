package repository

import (
	"context"
	"testing"

	"mcpdocs/models"
	"mcpdocs/utils"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)

	return db
}

func TestNewUserRepository(t *testing.T) {
	t.Run("should create new user repository", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		assert.NotNil(t, repo)
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		user := &models.User{
			ID:           "user-123",
			Email:        "test@example.com",
			UserName:     "Test User",
			PasswordHash: "hashed-password",
		}

		err := repo.CreateUser(ctx, user)
		assert.NoError(t, err)

		// Verify user was created
		var foundUser models.User
		err = db.First(&foundUser, "id = ?", user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		user1 := &models.User{
			ID:           "user-1",
			Email:        "duplicate@example.com",
			UserName:     "User 1",
			PasswordHash: "hashed-password",
		}

		err := repo.CreateUser(ctx, user1)
		assert.NoError(t, err)

		// Try to create another user with same email
		user2 := &models.User{
			ID:           "user-2",
			Email:        "duplicate@example.com",
			UserName:     "User 2",
			PasswordHash: "hashed-password",
		}

		err = repo.CreateUser(ctx, user2)
		assert.Error(t, err)
		assert.Equal(t, ErrUserAlreadyExists, err)
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	t.Run("should get user by email", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		user := &models.User{
			ID:           "user-123",
			Email:        "test@example.com",
			UserName:     "Test User",
			PasswordHash: "hashed-password",
		}

		err := repo.CreateUser(ctx, user)
		assert.NoError(t, err)

		foundUser, err := repo.GetUserByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
	})
}

func TestUserRepository_ValidateCredentials(t *testing.T) {
	t.Run("should validate correct credentials", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		password := "correctPassword123"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		user := &models.User{
			ID:           "user-123",
			Email:        "test@example.com",
			UserName:     "Test User",
			PasswordHash: hashedPassword,
		}

		err = repo.CreateUser(ctx, user)
		assert.NoError(t, err)

		validatedUser, err := repo.ValidateCredentials(ctx, user.Email, password)
		assert.NoError(t, err)
		assert.NotNil(t, validatedUser)
		assert.Equal(t, user.ID, validatedUser.ID)
	})

	t.Run("should return error for incorrect password", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		password := "correctPassword123"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		user := &models.User{
			ID:           "user-123",
			Email:        "test@example.com",
			UserName:     "Test User",
			PasswordHash: hashedPassword,
		}

		err = repo.CreateUser(ctx, user)
		assert.NoError(t, err)

		_, err = repo.ValidateCredentials(ctx, user.Email, "wrongPassword")
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
	})

	t.Run("should return error for nonexistent user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		_, err := repo.ValidateCredentials(ctx, "nonexistent@example.com", "somePassword")
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
	})
}
