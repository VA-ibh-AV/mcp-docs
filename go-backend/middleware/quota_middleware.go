package middleware

import (
	"net/http"

	"mcpdocs/models"
	"mcpdocs/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QuotaMiddleware checks if the user has exceeded their plan's usage limits.
func QuotaMiddleware(usageRepo *repository.UsageRepository, planRepo *repository.PlanRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDInterface.(string)

		plan := c.MustGet("plan").(*models.Plan)
		sub := c.MustGet("subs").(*models.Subcription)

		// Fetch user's usage for the current period
		periodStart := sub.PeriodStart
		periodEnd := sub.PeriodEnd

		usage, err := usageRepo.GetUsageByUserAndPeriod(userID, periodStart, periodEnd)
		if err != nil && err != gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage"})
			return
		}

		if usage == nil {
			usage = &models.Usage{
				UserID:       userID,
				UsageMetrics: models.PlanMetrics{},
				PeriodStart:  periodStart,
				PeriodEnd:    periodEnd,
			}
		}

		// Check if usage exceeds plan limits
		if usage.UsageMetrics.Projects >= plan.Metrics.Projects ||
			usage.UsageMetrics.SSEExecutions >= plan.Metrics.SSEExecutions ||
			usage.UsageMetrics.MaxIndexPerProject >= plan.Metrics.MaxIndexPerProject ||
			usage.UsageMetrics.AIPoweredSearchExecutions >= plan.Metrics.AIPoweredSearchExecutions {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{"error": "Quota exceeded"})
			return
		}

		c.Next()
	}
}
