package services

import (
	"context"
	"errors"
	"testing"

	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ValidateCredentials(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_RegisterUser(t *testing.T) {
	t.Run("should register user successfully", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		req := &schema.RegisterRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		// First call to GetUserByEmail returns error (user doesn't exist)
		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(nil, repository.ErrUserNotFound)
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		err := service.RegisterUser(ctx, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		req := &schema.RegisterRequest{
			Email:    "existing@example.com",
			Name:     "Existing User",
			Password: "password123",
		}

		existingUser := &models.User{
			ID:    "existing-user-id",
			Email: req.Email,
		}

		// GetUserByEmail returns existing user
		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(existingUser, nil)

		err := service.RegisterUser(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, repository.ErrUserAlreadyExists, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when CreateUser fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		req := &schema.RegisterRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(nil, repository.ErrUserNotFound)
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("database error"))

		err := service.RegisterUser(ctx, req)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_LoginUser(t *testing.T) {
	t.Run("should login user successfully with valid credentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		email := "test@example.com"
		password := "password123"

		user := &models.User{
			ID:    "user-id",
			Email: email,
		}

		mockRepo.On("ValidateCredentials", ctx, email, password).Return(user, nil)

		isValid, err := service.LoginUser(ctx, email, password)

		assert.NoError(t, err)
		assert.True(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error with invalid credentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		email := "test@example.com"
		password := "wrongpassword"

		mockRepo.On("ValidateCredentials", ctx, email, password).Return(nil, repository.ErrUserNotFound)

		isValid, err := service.LoginUser(ctx, email, password)

		assert.Error(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewAuthService(mockRepo)
		ctx := context.Background()

		email := "test@example.com"
		password := "password123"

		mockRepo.On("ValidateCredentials", ctx, email, password).Return(nil, errors.New("database error"))

		isValid, err := service.LoginUser(ctx, email, password)

		assert.Error(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})
}
