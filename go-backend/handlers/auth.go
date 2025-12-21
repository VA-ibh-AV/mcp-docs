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
	authService  *services.AuthService
	tokenService *services.TokenService
}

func NewAuthHandler(authService *services.AuthService, tokenService *services.TokenService) *AuthHandler {
	return &AuthHandler{authService: authService, tokenService: tokenService}
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

func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req schema.LoginRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		slog.ErrorContext(ctx, "Error binding body with JSON", "error", err)
		return
	}

	slog.InfoContext(ctx, "Request", "request", req)

	isValid, err := h.authService.LoginUser(ctx, req.Email, req.Password)
	if err != nil || !isValid {
		slog.ErrorContext(ctx, "Invalid credentials", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, refreshToken, err := h.tokenService.CreateTokens(ctx, req.Email)
	if err != nil {
		slog.ErrorContext(ctx, "Error creating tokens", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create tokens"})
		return
	}

	c.SetCookie("refresh_token", refreshToken, 7*24*60*60, "/", "", false, true) // 7 days expiry

	slog.InfoContext(ctx, "User Logged In", "email", req.Email)
	c.JSON(http.StatusOK, schema.TokenResponse{
		AccessToken: accessToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		slog.ErrorContext(ctx, "Refresh token cookie not found", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	newAccessToken, newRefreshToken, err := h.tokenService.RotateRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.ErrorContext(ctx, "Error rotating refresh token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.SetCookie("refresh_token", newRefreshToken, 7*24*60*60, "/", "", false, true) // 7 days expiry

	slog.InfoContext(ctx, "Refresh token rotated")
	c.JSON(http.StatusOK, schema.TokenResponse{
		AccessToken: newAccessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		slog.ErrorContext(ctx, "Refresh token cookie not found", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token not found"})
		return
	}

	err = h.tokenService.RevokeAllRefreshTokensForUser(ctx, refreshToken)
	if err != nil {
		slog.ErrorContext(ctx, "Error revoking refresh tokens", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke refresh tokens"})
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true) // Delete cookie

	slog.InfoContext(ctx, "User Logged Out")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
