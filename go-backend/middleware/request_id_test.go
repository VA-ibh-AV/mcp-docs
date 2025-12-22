package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should add request ID to context", func(t *testing.T) {
		var capturedRequestID string

		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", func(c *gin.Context) {
			rid := c.Request.Context().Value(RequestIDKey)
			if rid != nil {
				capturedRequestID = rid.(string)
			}
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedRequestID)
	})

	t.Run("should set X-Request-ID header", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})

	t.Run("should generate unique request IDs", func(t *testing.T) {
		var requestID1, requestID2 string

		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// First request
		req1 := httptest.NewRequest("GET", "/test", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		requestID1 = w1.Header().Get("X-Request-ID")

		// Second request
		req2 := httptest.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		requestID2 = w2.Header().Get("X-Request-ID")

		assert.NotEqual(t, requestID1, requestID2, "Each request should have a unique request ID")
	})

	t.Run("should allow next handler to execute", func(t *testing.T) {
		handlerCalled := false

		router := gin.New()
		router.Use(RequestID())
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Next handler should be called")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
