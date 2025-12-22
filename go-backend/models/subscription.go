package models

import "time"

type Subcription struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	UserID               string    `json:"user_id"` // user email
	PlanID               uint      `json:"plan_id"`
	Status               string    `json:"status"` // e.g., "active", "canceled", "past_due"
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
	StripeSubscriptionID string    `json:"stripe_subscription_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
