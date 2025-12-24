package main

import (
	"log/slog"
	"mcpdocs/app"
	main_app "mcpdocs/app"
	"mcpdocs/config"
	"mcpdocs/db"
	"mcpdocs/handlers"
	"mcpdocs/kafka"
	"mcpdocs/logger"
	"mcpdocs/middleware"
	"mcpdocs/models"
)

func main() {

	app := app.NewApp()
	config := config.LoadConfig()
	db, err := db.NewPostgres(config)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.Plan{}, &models.Subcription{}, &models.Usage{}, &models.Project{}, &models.IndexingRequest{}, &models.IndexingJob{})

	// Start Kafka Consumer for testing
	go func() {
		consumer, err := kafka.NewConsumer(config.KafkaBrokers)
		if err != nil {
			slog.Error("Failed to create Kafka consumer: " + err.Error())
			return
		}
		slog.Info("Starting Kafka consumer for indexing_requests")
		consumer.Consume("indexing_requests")
	}()

	container := main_app.NewContainer(db, config)
	authHandler := container.AuthHandler

	log := logger.New()
	slog.SetDefault(log)

	slog.Info("Starting application")

	app.Router.Use(middleware.RequestID())

	app.Router.GET("/", handlers.HealthCheck)
	auth := app.Router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	api := app.Router.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.POST("/projects", container.ProjectHandler.CreateProject)
		api.GET("/projects", container.ProjectHandler.GetProjects)
		api.GET("/projects/:id", container.ProjectHandler.GetProject)
		api.PUT("/projects/:id", container.ProjectHandler.UpdateProject)
		api.DELETE("/projects/:id", container.ProjectHandler.DeleteProject)

		api.GET("/plans", container.PlanHandler.ListPlans)
		api.GET("/plans/:id", container.PlanHandler.GetPlan)
		api.POST("/plans", container.PlanHandler.CreatePlan)

		api.POST("/subscription", container.SubscriptionHandler.CreateSubscription)
		api.GET("/subscription", container.SubscriptionHandler.GetSubscription)
		api.GET("/subscriptions", container.SubscriptionHandler.ListSubscriptions)
		api.POST("/subscription/:subscriptionID/cancel", container.SubscriptionHandler.CancelSubscription)

		api.POST("/indexing/requests", container.IndexingHandler.CreateIndexingRequest)
		api.GET("/indexing/requests/:id", container.IndexingHandler.GetIndexingRequest)
		api.PUT("/indexing/requests/:id/status", container.IndexingHandler.UpdateIndexingRequestStatus)
		api.GET("/projects/:id/indexing-requests", container.IndexingHandler.GetIndexingRequestsByProject)

		api.POST("/indexing/jobs", container.IndexingHandler.CreateIndexingJob)
		api.GET("/indexing/jobs/:id", container.IndexingHandler.GetIndexingJob)
		api.PUT("/indexing/jobs/:id/status", container.IndexingHandler.UpdateIndexingJobStatus)
		api.GET("/indexing/requests/:id/jobs", container.IndexingHandler.GetIndexingJobsByRequest)
	}

	app.Start()

}
