package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables from system")
	}

	port := os.Getenv("PORT")

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}

	log.Println("Server started on port: ", port)

}
