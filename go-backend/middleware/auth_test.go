package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mcpdocs/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired(t *testing.T) {
	// Set up test JWT secret
	originalSecret := os.Getenv("JWT_SECRET_KEY")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-auth-middleware")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET_KEY", originalSecret)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
	}()

	gin.SetMode(gin.TestMode)

	t.Run("should allow request with valid token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			userID, exists := c.Get("userID")
			assert.True(t, exists)
			c.JSON(200, gin.H{"userID": userID})
		})

		// Generate a valid token
		token, err := utils.GenerateJWT("test-user-123", 15*time.Minute)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject request without authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with invalid token format", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with missing Bearer prefix", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		token, err := utils.GenerateJWT("test-user", 15*time.Minute)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", token) // Missing "Bearer " prefix
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with invalid token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with expired token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Generate expired token
		token, err := utils.GenerateJWT("test-user", -1*time.Hour)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should set userID in context for valid token", func(t *testing.T) {
		expectedUserID := "user-set-in-context"
		var capturedUserID string

		router := gin.New()
		router.Use(AuthRequired())
		router.GET("/protected", func(c *gin.Context) {
			userID, exists := c.Get("userID")
			if exists {
				capturedUserID = userID.(string)
			}
			c.JSON(200, gin.H{"userID": userID})
		})

		token, err := utils.GenerateJWT(expectedUserID, 15*time.Minute)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedUserID, capturedUserID)
	})
}
