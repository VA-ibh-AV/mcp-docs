package middleware

import (
	"log/slog"
	"mcpdocs/utils"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header missing"})
			c.Abort()
			return
		}

		parts := len("Bearer ")
		if len(authHeader) <= parts || authHeader[:parts] != "Bearer " {
			slog.ErrorContext(ctx, "Invalid authorization header format", "header", authHeader)
			c.JSON(401, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := authHeader[parts:]
		userID, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
