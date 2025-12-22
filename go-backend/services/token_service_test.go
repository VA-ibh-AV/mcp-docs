package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, userID, token string) (*models.RefreshToken, error) {
	args := m.Called(ctx, userID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) FindValid(ctx context.Context, token string) (*models.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestTokenService_CreateTokens(t *testing.T) {
	t.Run("should create access and refresh tokens successfully", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		userID := "test-user-123"

		refreshToken := &models.RefreshToken{
			ID:        "token-id",
			UserID:    userID,
			TokenHash: "hashed-token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo.On("Create", ctx, userID, mock.AnythingOfType("string")).Return(refreshToken, nil)

		accessToken, refresh, err := service.CreateTokens(ctx, userID)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refresh)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		userID := "test-user-123"

		mockRepo.On("Create", ctx, userID, mock.AnythingOfType("string")).Return(nil, errors.New("database error"))

		_, _, err := service.CreateTokens(ctx, userID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTokenService_RotateRefreshToken(t *testing.T) {
	t.Run("should rotate refresh token successfully", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		oldToken := "old-refresh-token"
		userID := "test-user-123"

		hashedOldToken := utils.HashToken(oldToken)

		existingToken := &models.RefreshToken{
			ID:        "token-id",
			UserID:    userID,
			TokenHash: hashedOldToken,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		newToken := &models.RefreshToken{
			ID:        "new-token-id",
			UserID:    userID,
			TokenHash: "new-hashed-token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo.On("FindValid", ctx, hashedOldToken).Return(existingToken, nil)
		mockRepo.On("Revoke", ctx, hashedOldToken).Return(nil)
		mockRepo.On("Create", ctx, userID, mock.AnythingOfType("string")).Return(newToken, nil)

		newAccessToken, newRefreshToken, err := service.RotateRefreshToken(ctx, oldToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when old token is invalid", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		oldToken := "invalid-token"

		hashedOldToken := utils.HashToken(oldToken)

		mockRepo.On("FindValid", ctx, hashedOldToken).Return(nil, repository.ErrRefreshTokenNotFound)

		_, _, err := service.RotateRefreshToken(ctx, oldToken)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAccessToken, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when revoke fails", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		oldToken := "old-token"
		userID := "test-user-123"

		hashedOldToken := utils.HashToken(oldToken)

		existingToken := &models.RefreshToken{
			ID:        "token-id",
			UserID:    userID,
			TokenHash: hashedOldToken,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockRepo.On("FindValid", ctx, hashedOldToken).Return(existingToken, nil)
		mockRepo.On("Revoke", ctx, hashedOldToken).Return(errors.New("revoke error"))

		_, _, err := service.RotateRefreshToken(ctx, oldToken)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTokenService_RevokeRefreshToken(t *testing.T) {
	t.Run("should revoke refresh token successfully", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		token := "refresh-token"

		hashedToken := utils.HashToken(token)

		mockRepo.On("Revoke", ctx, hashedToken).Return(nil)

		err := service.RevokeRefreshToken(ctx, token)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when revoke fails", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		token := "refresh-token"

		hashedToken := utils.HashToken(token)

		mockRepo.On("Revoke", ctx, hashedToken).Return(errors.New("revoke error"))

		err := service.RevokeRefreshToken(ctx, token)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTokenService_RevokeAllRefreshTokensForUser(t *testing.T) {
	t.Run("should revoke all refresh tokens for user successfully", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		userID := "test-user-123"

		mockRepo.On("RevokeAllForUser", ctx, userID).Return(nil)

		err := service.RevokeAllRefreshTokensForUser(ctx, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when revoke all fails", func(t *testing.T) {
		mockRepo := new(MockRefreshTokenRepository)
		service := NewTokenService(mockRepo)
		ctx := context.Background()
		userID := "test-user-123"

		mockRepo.On("RevokeAllForUser", ctx, userID).Return(errors.New("revoke all error"))

		err := service.RevokeAllRefreshTokensForUser(ctx, userID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
