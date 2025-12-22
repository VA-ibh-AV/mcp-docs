package repository

import (
	"mcpdocs/models"
	"time"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) CreateSubscription(subscription *models.Subcription) error {
	return r.db.Create(subscription).Error
}

func (r *SubscriptionRepository) GetActiveSubscriptionByUserID(userID string, currentTime time.Time) (*models.Subcription, error) {
	var subscription models.Subcription
	err := r.db.Where("user_id = ? AND start_date <= ? AND end_date >= ?", userID, currentTime, currentTime).First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) UpdateSubscription(subscription *models.Subcription) error {
	return r.db.Save(subscription).Error
}

func (r *SubscriptionRepository) DeleteSubscription(subscriptionID uint) error {
	return r.db.Delete(&models.Subcription{}, subscriptionID).Error
}

func (r *SubscriptionRepository) ListSubscriptionsByUser(userID string) ([]models.Subcription, error) {
	var subscriptions []models.Subcription
	err := r.db.Where("user_id = ?", userID).Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *SubscriptionRepository) CancelSubscriptionByIDString(stripeSubscriptionID string) error {
	return r.db.Model(&models.Subcription{}).
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		Update("status", "canceled").Error
}

func (r *SubscriptionRepository) GetSubscriptionByStripeID(stripeSubscriptionID string) (*models.Subcription, error) {
	var subscription models.Subcription
	err := r.db.Where("stripe_subscription_id = ?", stripeSubscriptionID).First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}
