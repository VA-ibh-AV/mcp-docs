package utils

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	// Set up test JWT secret
	originalSecret := os.Getenv("JWT_SECRET_KEY")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-testing")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET_KEY", originalSecret)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
	}()

	t.Run("should generate valid JWT token", func(t *testing.T) {
		userID := "user-123"
		token, err := GenerateJWT(userID, 15*time.Minute)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("should generate different tokens for different users", func(t *testing.T) {
		token1, err1 := GenerateJWT("user-1", 15*time.Minute)
		token2, err2 := GenerateJWT("user-2", 15*time.Minute)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, token1, token2, "Different users should get different tokens")
	})

	t.Run("should generate token with expiry", func(t *testing.T) {
		userID := "user-123"
		duration := 1 * time.Hour
		token, err := GenerateJWT(userID, duration)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("should include user_id in token", func(t *testing.T) {
		userID := "test-user-456"
		token, err := GenerateJWT(userID, 15*time.Minute)
		
		assert.NoError(t, err)
		
		// Validate and extract userID
		extractedUserID, err := ValidateJWT(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, extractedUserID)
	})

	t.Run("should handle empty user ID", func(t *testing.T) {
		userID := ""
		token, err := GenerateJWT(userID, 15*time.Minute)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("should handle very short expiry", func(t *testing.T) {
		userID := "user-123"
		token, err := GenerateJWT(userID, 1*time.Millisecond)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestValidateJWT(t *testing.T) {
	// Set up test JWT secret
	originalSecret := os.Getenv("JWT_SECRET_KEY")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-testing")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET_KEY", originalSecret)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
	}()

	t.Run("should validate correct token", func(t *testing.T) {
		userID := "user-789"
		token, err := GenerateJWT(userID, 15*time.Minute)
		assert.NoError(t, err)
		
		extractedUserID, err := ValidateJWT(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, extractedUserID)
	})

	t.Run("should reject invalid token format", func(t *testing.T) {
		invalidToken := "this-is-not-a-valid-jwt-token"
		
		_, err := ValidateJWT(invalidToken)
		assert.Error(t, err)
	})

	t.Run("should reject empty token", func(t *testing.T) {
		_, err := ValidateJWT("")
		assert.Error(t, err)
	})

	t.Run("should reject expired token", func(t *testing.T) {
		userID := "user-expired"
		// Create token that expires immediately
		token, err := GenerateJWT(userID, -1*time.Hour)
		assert.NoError(t, err)
		
		_, err = ValidateJWT(token)
		assert.Error(t, err, "Expired token should be rejected")
	})

	t.Run("should reject token with wrong secret", func(t *testing.T) {
		userID := "user-123"
		token, err := GenerateJWT(userID, 15*time.Minute)
		assert.NoError(t, err)
		
		// Change the secret
		os.Setenv("JWT_SECRET_KEY", "different-secret-key")
		
		_, err = ValidateJWT(token)
		assert.Error(t, err, "Token signed with different secret should be rejected")
	})
}
