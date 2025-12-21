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

	db.AutoMigrate(&models.User{}, &models.RefreshToken{})

	container := main_app.NewContainer(db)
	authHandler := container.AuthHandler

	log := logger.New()
	slog.SetDefault(log)

	slog.Info("Starting application")

	app.Router.Use(middleware.RequestID())

	app.Router.GET("/", handlers.HealthCheck)
	app.Router.POST("/auth/register", authHandler.Register)
	app.Router.POST("/auth/login", authHandler.Login)
	app.Router.POST("/auth/refresh", authHandler.Refresh)
	app.Router.POST("/auth/logout", authHandler.Logout)
	app.Start()

}
