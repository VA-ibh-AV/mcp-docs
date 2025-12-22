package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashToken(t *testing.T) {
	t.Run("should hash token consistently", func(t *testing.T) {
		token := "test-token-123"
		hash1 := HashToken(token)
		hash2 := HashToken(token)
		
		assert.NotEmpty(t, hash1)
		assert.Equal(t, hash1, hash2, "Same token should produce same hash")
	})

	t.Run("should produce different hashes for different tokens", func(t *testing.T) {
		token1 := "token-1"
		token2 := "token-2"
		
		hash1 := HashToken(token1)
		hash2 := HashToken(token2)
		
		assert.NotEqual(t, hash1, hash2, "Different tokens should produce different hashes")
	})

	t.Run("should produce hex encoded hash", func(t *testing.T) {
		token := "test-token"
		hash := HashToken(token)
		
		// SHA256 produces 32 bytes, hex encoding makes it 64 characters
		assert.Equal(t, 64, len(hash), "SHA256 hash should be 64 characters in hex")
	})

	t.Run("should handle empty token", func(t *testing.T) {
		token := ""
		hash := HashToken(token)
		
		assert.NotEmpty(t, hash, "Empty token should still produce a hash")
		assert.Equal(t, 64, len(hash))
	})
}
