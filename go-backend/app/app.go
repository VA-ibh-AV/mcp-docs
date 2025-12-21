package app

import (
	"github.com/gin-gonic/gin"
	"mcpdocs/config"
	"log"
)

type App struct {
	Router *gin.Engine
	Config *config.Config
}

func NewApp() *App {
	return &App{
		Router: gin.Default(),
		Config: config.LoadConfig(),
	}
}

func (app *App) Start() {
	log.Println("Server started on port: ", app.Config.Port)
	app.Router.Run(":" + app.Config.Port)
}
