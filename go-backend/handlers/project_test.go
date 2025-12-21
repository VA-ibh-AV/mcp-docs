package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewProjectHandler(t *testing.T) {
	t.Run("should create new project handler", func(t *testing.T) {
		handler := NewProjectHandler()
		assert.NotNil(t, handler)
	})
}

func TestProjectHandler_CreateProject(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return not implemented status", func(t *testing.T) {
		handler := NewProjectHandler()
		router := gin.New()
		router.POST("/projects", handler.CreateProject)

		req := httptest.NewRequest("POST", "/projects", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
		assert.Contains(t, w.Body.String(), "not implemented")
	})
}
