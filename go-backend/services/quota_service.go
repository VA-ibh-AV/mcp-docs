package services

import (
	"mcpdocs/models"
	"mcpdocs/repository"
	"time"

	"gorm.io/gorm"
)

type QuotaService struct {
	usageRepo *repository.UsageRepository
}

func NewQuotaService(usageRepo *repository.UsageRepository) *QuotaService {
	return &QuotaService{usageRepo: usageRepo}
}

func (s *QuotaService) CheckAndUpdateUsage(userID string, additionalUsage models.PlanMetrics) (bool, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	usage, err := s.usageRepo.GetUsageByUserAndPeriod(userID, periodStart, periodEnd)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			usage = &models.Usage{
				UserID:       userID,
				UsageMetrics: models.PlanMetrics{},
				PeriodStart:  periodStart,
				PeriodEnd:    periodEnd,
			}
		} else {
			return false, err
		}
	}

	// Update usage metrics
	usage.UsageMetrics.Projects += additionalUsage.Projects
	usage.UsageMetrics.SSEExecutions += additionalUsage.SSEExecutions
	usage.UsageMetrics.MaxIndexPerProject += additionalUsage.MaxIndexPerProject
	usage.UsageMetrics.AIPoweredSearchExecutions += additionalUsage.AIPoweredSearchExecutions

	if usage.ID == 0 {
		err = s.usageRepo.CreateUsage(usage)
	} else {
		err = s.usageRepo.UpdateUsage(usage)
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
