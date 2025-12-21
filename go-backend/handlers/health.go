package handlers

import (
	"github.com/gin-gonic/gin"
)

type HealthCheckResponse struct {
	Status string `json:"status" example:"OK"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, HealthCheckResponse{Status: "OK"})
}