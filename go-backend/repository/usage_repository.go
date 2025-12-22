package repository

import (
	"mcpdocs/models"
	"time"

	"gorm.io/gorm"
)

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) CreateUsage(usage *models.Usage) error {
	return r.db.Create(usage).Error
}

func (r *UsageRepository) GetUsageByUserAndPeriod(userID string, periodStart, periodEnd time.Time) (*models.Usage, error) {
	var usage models.Usage
	err := r.db.Where("user_id = ? AND period_start = ? AND period_end = ?", userID, periodStart, periodEnd).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *UsageRepository) UpdateUsage(usage *models.Usage) error {
	return r.db.Save(usage).Error
}

func (r *UsageRepository) DeleteUsage(usageID uint) error {
	return r.db.Delete(&models.Usage{}, usageID).Error
}

func (r *UsageRepository) ListUsagesByUser(userID string) ([]models.Usage, error) {
	var usages []models.Usage
	err := r.db.Where("user_id = ?", userID).Find(&usages).Error
	if err != nil {
		return nil, err
	}
	return usages, nil
}
