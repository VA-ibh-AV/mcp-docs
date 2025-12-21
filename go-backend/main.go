package main

import (
	"log/slog"
	"mcpdocs/app"
	"mcpdocs/config"
	"mcpdocs/db"
	"mcpdocs/handlers"
	"mcpdocs/logger"
	"mcpdocs/middleware"
	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/services"
)

func main() {
	app := app.NewApp()
	config := config.LoadConfig()
	db, err := db.NewPostgres(config)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	db.AutoMigrate(&models.User{})

	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	log := logger.New()
	slog.SetDefault(log)

	slog.Info("Starting application")

	app.Router.Use(middleware.RequestID())

	app.Router.GET("/", handlers.HealthCheck)
	app.Router.POST("/auth/register", authHandler.Register)
	app.Start()

}
