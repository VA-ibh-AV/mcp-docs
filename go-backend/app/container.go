package app

import (
	"mcpdocs/handlers"
	"mcpdocs/repository"
	"mcpdocs/services"

	"gorm.io/gorm"
)

type Contaner struct {
	AuthHandler *handlers.AuthHandler
}

func NewContainer(db *gorm.DB) *Contaner {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)

	// Services
	authService := services.NewAuthService(userRepo)
	tokenService := services.NewTokenService(tokenRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, tokenService)

	return &Contaner{
		AuthHandler: authHandler,
	}
}
