package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("should hash password successfully", func(t *testing.T) {
		password := "test-password-123"
		hash, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash, "Hash should be different from password")
	})

	t.Run("should produce different hashes for same password", func(t *testing.T) {
		password := "same-password"
		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "bcrypt should produce different hashes with different salts")
	})

	t.Run("should handle empty password", func(t *testing.T) {
		password := ""
		hash, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func TestCheckPasswordHash(t *testing.T) {
	t.Run("should return true for correct password", func(t *testing.T) {
		password := "correct-password"
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		
		isValid := CheckPasswordHash(password, hash)
		assert.True(t, isValid, "Correct password should validate")
	})

	t.Run("should return false for incorrect password", func(t *testing.T) {
		password := "correct-password"
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		
		isValid := CheckPasswordHash("wrong-password", hash)
		assert.False(t, isValid, "Incorrect password should not validate")
	})

	t.Run("should return false for invalid hash", func(t *testing.T) {
		password := "some-password"
		invalidHash := "not-a-valid-bcrypt-hash"
		
		isValid := CheckPasswordHash(password, invalidHash)
		assert.False(t, isValid, "Invalid hash should not validate")
	})

	t.Run("should handle empty password", func(t *testing.T) {
		password := ""
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		
		isValid := CheckPasswordHash("", hash)
		assert.True(t, isValid, "Empty password should validate against its hash")
	})
}
