package repository

import (
	"context"
	"errors"
	"mcpdocs/models"
	"mcpdocs/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, token string) (*models.RefreshToken, error) {
	hashedToken := utils.HashToken(token)
	refreshToken := &models.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hashedToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
	}

	if err := r.db.WithContext(ctx).Create(refreshToken).Error; err != nil {
		return nil, err
	}
	return refreshToken, nil
}

func (r *RefreshTokenRepository) FindValid(ctx context.Context, token string) (*models.RefreshToken, error) {
	hashedToken := utils.HashToken(token)
	var refreshToken models.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND expires_at > ? AND revoked = false", hashedToken, time.Now()).First(&refreshToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	hashedToken := utils.HashToken(token)
	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("token_hash = ?", hashedToken).Update("revoked", true)
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return result.Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true)
	return result.Error
}
