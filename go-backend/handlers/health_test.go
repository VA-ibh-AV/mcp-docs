package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK status", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return OK status in response body", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), `"status":"OK"`)
	})

	t.Run("should return valid JSON response", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})
}
