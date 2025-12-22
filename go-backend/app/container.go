package app

import (
	"mcpdocs/handlers"
	"mcpdocs/repository"
	"mcpdocs/services"

	"gorm.io/gorm"
)

type Contaner struct {
	AuthHandler         *handlers.AuthHandler
	ProjectHandler      *handlers.ProjectHandler
	PlanHandler         *handlers.PlanHandler
	SubscriptionHandler *handlers.SubscriptionHandler
}

func NewContainer(db *gorm.DB) *Contaner {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	planRepo := repository.NewPlanRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)

	// Services
	authService := services.NewAuthService(userRepo)
	tokenService := services.NewTokenService(tokenRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, tokenService)
	projectHandler := handlers.NewProjectHandler()
	planHandler := handlers.NewPlanHandler(planRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionRepo, planRepo)

	return &Contaner{
		AuthHandler:         authHandler,
		ProjectHandler:      projectHandler,
		PlanHandler:         planHandler,
		SubscriptionHandler: subscriptionHandler,
	}
}
