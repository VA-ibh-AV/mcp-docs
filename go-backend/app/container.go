package app

import (
	"log"
	"mcpdocs/config"
	"mcpdocs/handlers"
	"mcpdocs/kafka"
	"mcpdocs/repository"
	"mcpdocs/services"

	"gorm.io/gorm"
)

type Contaner struct {
	AuthHandler         *handlers.AuthHandler
	ProjectHandler      *handlers.ProjectHandler
	PlanHandler         *handlers.PlanHandler
	SubscriptionHandler *handlers.SubscriptionHandler
	IndexingHandler     *handlers.IndexingHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Contaner {
	// Kafka
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Printf("Failed to create Kafka producer: %v", err)
		// Continue without Kafka for now, or handle error appropriately
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	planRepo := repository.NewPlanRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	indexingRequestRepo := repository.NewIndexingRequestRepository(db)
	indexingJobRepo := repository.NewIndexingJobRepository(db)

	// Services
	authService := services.NewAuthService(userRepo)
	tokenService := services.NewTokenService(tokenRepo)
	projectService := services.NewProjectService(projectRepo)
	indexingService := services.NewIndexingService(indexingRequestRepo, indexingJobRepo, producer)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, tokenService)
	projectHandler := handlers.NewProjectHandler(projectService)
	planHandler := handlers.NewPlanHandler(planRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionRepo, planRepo)
	indexingHandler := handlers.NewIndexingHandler(indexingService)

	return &Contaner{
		AuthHandler:         authHandler,
		ProjectHandler:      projectHandler,
		PlanHandler:         planHandler,
		SubscriptionHandler: subscriptionHandler,
		IndexingHandler:     indexingHandler,
	}
}
