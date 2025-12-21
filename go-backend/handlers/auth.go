package handlers

import (
	"errors"
	"log/slog"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"mcpdocs/services"

	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req schema.RegisterRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		slog.ErrorContext(ctx, "Error binding body with JSON", "error", err)
		return
	}

	slog.InfoContext(ctx, "Request", "request", req)

	if err := h.authService.RegisterUser(ctx, &req); err != nil {

		if errors.Is(err, repository.ErrUserAlreadyExists) {
			c.JSON(409, gin.H{"error": "user already exists"})
			return
		}

		slog.ErrorContext(ctx, "Error registering user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.InfoContext(ctx, "User Registered", "email", req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
