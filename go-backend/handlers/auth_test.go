package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"mcpdocs/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx context.Context, req *schema.RegisterRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) LoginUser(ctx context.Context, email, password string) (bool, error) {
	args := m.Called(ctx, email, password)
	return args.Bool(0), args.Error(1)
}

// MockTokenService is a mock implementation of TokenService
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) CreateTokens(ctx context.Context, userID string) (string, string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) RotateRefreshToken(ctx context.Context, oldToken string) (string, string, error) {
	args := m.Called(ctx, oldToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenService) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestNewAuthHandler(t *testing.T) {
	t.Run("should create new auth handler", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockTokenService := new(MockTokenService)

		// We need to cast to the actual service types
		// For this to work, we need to modify the handler or create a factory
		authService := &services.AuthService{}
		tokenService := &services.TokenService{}

		handler := NewAuthHandler(authService, tokenService)
		assert.NotNil(t, handler)

		// Cleanup
		mockAuthService.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
	})
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should register user successfully", func(t *testing.T) {
		// Create mocks
		mockUserRepo := new(MockUserRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := &services.TokenService{}

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/register", handler.Register)

		registerReq := schema.RegisterRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		// Mock expectations
		mockUserRepo.On("GetUserByEmail", mock.Anything, registerReq.Email).Return(nil, repository.ErrUserNotFound)
		mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User registered successfully")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid request body", func(t *testing.T) {
		authService := &services.AuthService{}
		tokenService := &services.TokenService{}
		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/register", handler.Register)

		invalidBody := []byte(`{"email": "invalid"}`)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return conflict when user already exists", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := &services.TokenService{}

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/register", handler.Register)

		registerReq := schema.RegisterRequest{
			Email:    "existing@example.com",
			Name:     "Existing User",
			Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		existingUser := &models.User{
			ID:    "existing-id",
			Email: registerReq.Email,
		}

		mockUserRepo.On("GetUserByEmail", mock.Anything, registerReq.Email).Return(existingUser, nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "user already exists")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when create user fails with database error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := &services.TokenService{}

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/register", handler.Register)

		registerReq := schema.RegisterRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		mockUserRepo.On("GetUserByEmail", mock.Anything, registerReq.Email).Return(nil, repository.ErrUserNotFound)
		mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("database error"))

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should login user successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/login", handler.Login)

		loginReq := schema.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)

		user := &models.User{
			ID:    "user-id",
			Email: loginReq.Email,
		}

		mockUserRepo.On("ValidateCredentials", mock.Anything, loginReq.Email, loginReq.Password).Return(user, nil)
		mockTokenRepo.On("Create", mock.Anything, loginReq.Email, mock.AnythingOfType("string")).Return(&models.RefreshToken{}, nil)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
		mockUserRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid credentials", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/login", handler.Login)

		loginReq := schema.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		mockUserRepo.On("ValidateCredentials", mock.Anything, loginReq.Email, loginReq.Password).Return(nil, repository.ErrUserNotFound)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid credentials")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid request body", func(t *testing.T) {
		authService := &services.AuthService{}
		tokenService := &services.TokenService{}
		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/login", handler.Login)

		invalidBody := []byte(`{"email": "invalid"}`)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when token creation fails", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := services.NewAuthService(mockUserRepo)
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/login", handler.Login)

		loginReq := schema.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)

		user := &models.User{
			ID:    "user-id",
			Email: loginReq.Email,
		}

		mockUserRepo.On("ValidateCredentials", mock.Anything, loginReq.Email, loginReq.Password).Return(user, nil)
		mockTokenRepo.On("Create", mock.Anything, loginReq.Email, mock.AnythingOfType("string")).Return(nil, errors.New("token creation error"))

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should refresh token successfully", func(t *testing.T) {
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := &services.AuthService{}
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/refresh", handler.Refresh)

		oldToken := "old-refresh-token"
		
		mockTokenRepo.On("FindValid", mock.Anything, mock.AnythingOfType("string")).Return(&models.RefreshToken{UserID: "user-id"}, nil)
		mockTokenRepo.On("Revoke", mock.Anything, mock.AnythingOfType("string")).Return(nil)
		mockTokenRepo.On("Create", mock.Anything, "user-id", mock.AnythingOfType("string")).Return(&models.RefreshToken{}, nil)

		req := httptest.NewRequest("POST", "/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: oldToken,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("should return error when refresh token cookie is missing", func(t *testing.T) {
		authService := &services.AuthService{}
		tokenService := &services.TokenService{}
		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/refresh", handler.Refresh)

		req := httptest.NewRequest("POST", "/refresh", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "refresh token not found")
	})

	t.Run("should return error when token rotation fails", func(t *testing.T) {
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := &services.AuthService{}
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/refresh", handler.Refresh)

		oldToken := "invalid-token"

		mockTokenRepo.On("FindValid", mock.Anything, mock.AnythingOfType("string")).Return(nil, repository.ErrRefreshTokenNotFound)

		req := httptest.NewRequest("POST", "/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: oldToken,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockTokenRepo.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should logout user successfully", func(t *testing.T) {
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := &services.AuthService{}
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/logout", handler.Logout)

		refreshToken := "refresh-token"

		mockTokenRepo.On("RevokeAllForUser", mock.Anything, refreshToken).Return(nil)

		req := httptest.NewRequest("POST", "/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: refreshToken,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Logged out successfully")
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("should return error when refresh token cookie is missing", func(t *testing.T) {
		authService := &services.AuthService{}
		tokenService := &services.TokenService{}
		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "refresh token not found")
	})

	t.Run("should return error when token revocation fails", func(t *testing.T) {
		mockTokenRepo := new(MockRefreshTokenRepository)
		authService := &services.AuthService{}
		tokenService := services.NewTokenService(mockTokenRepo)

		handler := NewAuthHandler(authService, tokenService)

		router := gin.New()
		router.POST("/logout", handler.Logout)

		refreshToken := "refresh-token"

		mockTokenRepo.On("RevokeAllForUser", mock.Anything, refreshToken).Return(errors.New("revocation error"))

		req := httptest.NewRequest("POST", "/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: refreshToken,
		})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockTokenRepo.AssertExpectations(t)
	})
}
