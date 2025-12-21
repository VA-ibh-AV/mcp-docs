package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJWTSecretKey(t *testing.T) {
	t.Run("should return JWT secret key from environment", func(t *testing.T) {
		expectedSecret := "test-jwt-secret-key"
		os.Setenv("JWT_SECRET_KEY", expectedSecret)
		defer os.Unsetenv("JWT_SECRET_KEY")

		secretKey := GetJWTSecretKey()

		assert.Equal(t, expectedSecret, secretKey)
	})

	t.Run("should return empty string when JWT_SECRET_KEY is not set", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET_KEY")

		secretKey := GetJWTSecretKey()

		assert.Equal(t, "", secretKey)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("should load config from environment variables", func(t *testing.T) {
		// Set up environment variables
		os.Setenv("PORT", "8080")
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "testuser")
		os.Setenv("POSTGRES_PASSWORD", "testpass")
		os.Setenv("POSTGRES_DB", "testdb")

		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("POSTGRES_HOST")
			os.Unsetenv("POSTGRES_PORT")
			os.Unsetenv("POSTGRES_USER")
			os.Unsetenv("POSTGRES_PASSWORD")
			os.Unsetenv("POSTGRES_DB")
		}()

		config := LoadConfig()

		assert.Equal(t, "8080", config.Port)
		assert.Equal(t, "localhost", config.PostgresHost)
		assert.Equal(t, "5432", config.PostgresPort)
		assert.Equal(t, "testuser", config.PostgresUser)
		assert.Equal(t, "testpass", config.PostgresPassword)
		assert.Equal(t, "testdb", config.PostgresDb)
	})

	t.Run("should handle missing environment variables", func(t *testing.T) {
		// Unset all environment variables
		os.Unsetenv("PORT")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DB")

		config := LoadConfig()

		// Config should be created with empty values
		assert.NotNil(t, config)
		assert.Equal(t, "", config.Port)
		assert.Equal(t, "", config.PostgresHost)
	})
}
