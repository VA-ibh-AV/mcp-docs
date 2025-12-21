package services

import (
	"context"
	"errors"
	"mcpdocs/repository"
	"mcpdocs/utils"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type TokenService struct {
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewTokenService(refreshTokenRepo *repository.RefreshTokenRepository) *TokenService {
	return &TokenService{refreshTokenRepo: refreshTokenRepo}
}

func (s *TokenService) CreateTokens(ctx context.Context, userID string) (accessToken string, refreshToken string, err error) {
	accessToken, err = utils.GenerateJWT(userID, 15*time.Minute) // 15 minutes
	if err != nil {
		return "", "", err
	}

	refresh := uuid.NewString()
	hash := utils.HashToken(refresh)

	_, err = s.refreshTokenRepo.Create(ctx, userID, hash)
	if err != nil {
		return "", "", err
	}

	return accessToken, refresh, nil
}

func (s *TokenService) RotateRefreshToken(ctx context.Context, oldToken string) (newAccessToken string, newRefreshToken string, err error) {
	refreshTokenRecord, err := s.refreshTokenRepo.FindValid(ctx, utils.HashToken(oldToken))
	if err != nil {
		return "", "", ErrInvalidAccessToken
	}

	err = s.refreshTokenRepo.Revoke(ctx, utils.HashToken(oldToken))
	if err != nil {
		return "", "", err
	}

	return s.CreateTokens(ctx, refreshTokenRecord.UserID)
}

func (s *TokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.refreshTokenRepo.Revoke(ctx, utils.HashToken(token))
}

func (s *TokenService) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	return s.refreshTokenRepo.RevokeAllForUser(ctx, userID)
}
