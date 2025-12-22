package schema

import (
	"time"
)

type CreateSubscriptionInput struct {
	UserID string `json:"user_id" binding:"required"`
	PlanID uint   `json:"plan_id" binding:"required"`
}

type UpdateSubscriptionInput struct {
	UserID string `json:"user_id"`
	PlanID uint   `json:"plan_id"`
	Status string `json:"status"`
}

type SubscriptionResponse struct {
	ID                   uint      `json:"id"`
	UserID               string    `json:"user_id"`
	PlanID               uint      `json:"plan_id"`
	Status               string    `json:"status"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
	StripeSubscriptionID string    `json:"stripe_subscription_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type SubscriptionRenewalWebhook struct {
	UserID               string    `json:"user_id" binding:"required"`
	StripeSubscriptionID string    `json:"stripe_subscription_id" binding:"required"`
	NewPeriodStart       time.Time `json:"new_period_start" binding:"required"`
	NewPeriodEnd         time.Time `json:"new_period_end" binding:"required"`
}

type SubscriptionCancellationWebhook struct {
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
}
