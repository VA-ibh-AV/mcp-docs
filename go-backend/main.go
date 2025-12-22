package main

import (
	"log/slog"
	"mcpdocs/app"
	main_app "mcpdocs/app"
	"mcpdocs/config"
	"mcpdocs/db"
	"mcpdocs/handlers"
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

	db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.Plan{}, &models.Subcription{}, &models.Usage{})

	container := main_app.NewContainer(db)
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
		api.GET("/project", container.ProjectHandler.CreateProject)
		api.GET("/plans", container.PlanHandler.ListPlans)
		api.GET("/plans/:id", container.PlanHandler.GetPlan)
		api.POST("/plans", container.PlanHandler.CreatePlan)

		api.POST("/subscription", container.SubscriptionHandler.CreateSubscription)
		api.GET("/subscription", container.SubscriptionHandler.GetSubscription)
		api.GET("/subscriptions", container.SubscriptionHandler.ListSubscriptions)
		api.POST("/subscription/:subscriptionID/cancel", container.SubscriptionHandler.CancelSubscription)
	}

	app.Start()

}
