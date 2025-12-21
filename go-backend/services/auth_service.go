package services

import (
	"context"
	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"mcpdocs/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepository *repository.UserRepository
}

func NewAuthService(userRepository *repository.UserRepository) *AuthService {
	return &AuthService{userRepository: userRepository}
}

func (s *AuthService) RegisterUser(ctx context.Context, handlerReq *schema.RegisterRequest) error {
	hash, err := utils.HashPassword(handlerReq.Password)
	if err != nil {
		return err
	}

	if _, err := s.userRepository.GetUserByEmail(ctx, handlerReq.Email); err == nil {
		return repository.ErrUserAlreadyExists
	}

	return s.userRepository.CreateUser(ctx, &models.User{
		ID:           uuid.NewString(),
		Email:        handlerReq.Email,
		UserName:     handlerReq.Name,
		PasswordHash: hash,
	})
}

func (s *AuthService) LoginUser(ctx context.Context, userID string, password string) (isValid bool, err error) {
	_, err = s.userRepository.ValidateCredentials(ctx, userID, password)
	if err != nil {
		return false, err
	}
	return true, nil
}
