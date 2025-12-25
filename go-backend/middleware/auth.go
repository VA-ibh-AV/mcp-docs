package middleware

import (
	"log/slog"
	"mcpdocs/utils"
	"os"

	"github.com/gin-gonic/gin"
)

// Internal service API key for service-to-service communication
var internalAPIKey = os.Getenv("INTERNAL_API_KEY")

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Check for internal service API key first (X-Internal-API-Key header)
		apiKey := c.GetHeader("X-Internal-API-Key")
		if apiKey != "" && internalAPIKey != "" && apiKey == internalAPIKey {
			// Internal service authenticated - set a system user ID
			c.Set("userID", "internal-service")
			c.Next()
			return
		}
		
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
