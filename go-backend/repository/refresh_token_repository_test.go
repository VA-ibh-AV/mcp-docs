package repository

import (
	"context"
	"testing"
	"time"

	"mcpdocs/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRefreshTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.RefreshToken{})
	assert.NoError(t, err)

	return db
}

func TestNewRefreshTokenRepository(t *testing.T) {
	t.Run("should create new refresh token repository", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		assert.NotNil(t, repo)
	})
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	t.Run("should create refresh token successfully", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"
		token := "test-token-123"

		refreshToken, err := repo.Create(ctx, userID, token)
		assert.NoError(t, err)
		assert.NotNil(t, refreshToken)
		assert.Equal(t, userID, refreshToken.UserID)
		assert.NotEmpty(t, refreshToken.ID)
		assert.NotEmpty(t, refreshToken.TokenHash)
	})

	t.Run("should hash token before storing", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"
		token := "test-token-456"

		refreshToken, err := repo.Create(ctx, userID, token)
		assert.NoError(t, err)
		assert.NotEqual(t, token, refreshToken.TokenHash, "Token should be hashed")
	})

	t.Run("should set expiry to 7 days", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"
		token := "test-token-789"

		before := time.Now().Add(7 * 24 * time.Hour)
		refreshToken, err := repo.Create(ctx, userID, token)
		after := time.Now().Add(7 * 24 * time.Hour)

		assert.NoError(t, err)
		assert.True(t, refreshToken.ExpiresAt.After(before.Add(-1*time.Second)))
		assert.True(t, refreshToken.ExpiresAt.Before(after.Add(1*time.Second)))
	})
}

func TestRefreshTokenRepository_FindValid(t *testing.T) {
	t.Run("should find valid refresh token", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"
		token := "test-token-123"

		createdToken, err := repo.Create(ctx, userID, token)
		assert.NoError(t, err)

		foundToken, err := repo.FindValid(ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, foundToken)
		assert.Equal(t, createdToken.ID, foundToken.ID)
		assert.Equal(t, userID, foundToken.UserID)
	})

	t.Run("should return error for expired token", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		// Create a token that's already expired
		expiredToken := &models.RefreshToken{
			ID:        "expired-token-id",
			UserID:    "user-123",
			TokenHash: "hashed-token",
			Revoked:   false,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}

		err := db.Create(expiredToken).Error
		assert.NoError(t, err)

		_, err = repo.FindValid(ctx, "original-token")
		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
	})

	t.Run("should return error for revoked token", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		// Create a revoked token
		revokedToken := &models.RefreshToken{
			ID:        "revoked-token-id",
			UserID:    "user-123",
			TokenHash: "hashed-token-revoked",
			Revoked:   true,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			CreatedAt: time.Now(),
		}

		err := db.Create(revokedToken).Error
		assert.NoError(t, err)

		_, err = repo.FindValid(ctx, "original-token")
		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
	})

	t.Run("should return error for nonexistent token", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		_, err := repo.FindValid(ctx, "nonexistent-token")
		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
	})
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	t.Run("should revoke refresh token successfully", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"
		token := "test-token-123"

		_, err := repo.Create(ctx, userID, token)
		assert.NoError(t, err)

		err = repo.Revoke(ctx, token)
		assert.NoError(t, err)

		// Verify token is revoked
		_, err = repo.FindValid(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
	})

	t.Run("should return error when revoking nonexistent token", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		err := repo.Revoke(ctx, "nonexistent-token")
		assert.Error(t, err)
		assert.Equal(t, ErrRefreshTokenNotFound, err)
	})
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	t.Run("should revoke all refresh tokens for user", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID := "user-123"

		// Create multiple tokens for the user
		_, err := repo.Create(ctx, userID, "token-1")
		assert.NoError(t, err)
		_, err = repo.Create(ctx, userID, "token-2")
		assert.NoError(t, err)
		_, err = repo.Create(ctx, userID, "token-3")
		assert.NoError(t, err)

		// Revoke all tokens for the user
		err = repo.RevokeAllForUser(ctx, userID)
		assert.NoError(t, err)

		// Verify all tokens are revoked
		_, err = repo.FindValid(ctx, "token-1")
		assert.Error(t, err)
		_, err = repo.FindValid(ctx, "token-2")
		assert.Error(t, err)
		_, err = repo.FindValid(ctx, "token-3")
		assert.Error(t, err)
	})

	t.Run("should not error when revoking for user with no tokens", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		err := repo.RevokeAllForUser(ctx, "user-with-no-tokens")
		assert.NoError(t, err)
	})

	t.Run("should not affect tokens of other users", func(t *testing.T) {
		db := setupRefreshTokenTestDB(t)
		repo := NewRefreshTokenRepository(db)
		ctx := context.Background()

		userID1 := "user-1"
		userID2 := "user-2"

		// Create tokens for different users
		_, err := repo.Create(ctx, userID1, "token-user-1")
		assert.NoError(t, err)
		_, err = repo.Create(ctx, userID2, "token-user-2")
		assert.NoError(t, err)

		// Revoke all tokens for user 1
		err = repo.RevokeAllForUser(ctx, userID1)
		assert.NoError(t, err)

		// Verify user 1's token is revoked
		_, err = repo.FindValid(ctx, "token-user-1")
		assert.Error(t, err)

		// Verify user 2's token is still valid
		foundToken, err := repo.FindValid(ctx, "token-user-2")
		assert.NoError(t, err)
		assert.NotNil(t, foundToken)
	})
}
