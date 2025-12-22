package middleware

import (
	"net/http"
	"time"

	"mcpdocs/repository"

	"github.com/gin-gonic/gin"
)

func SubscriptionMiddleware(subRepo *repository.SubscriptionRepository, planRepo *repository.PlanRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDInterface.(string)

		sub, err := subRepo.GetActiveSubscriptionByUserID(userID, time.Now())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
			return
		}
		if sub == nil {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{"error": "No active subscription"})
			return
		}

		plan, err := planRepo.GetPlanByID(sub.PlanID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plan"})
			return
		}

		c.Set("subs", sub)
		c.Set("plan", plan)

		c.Next()
	}
}
