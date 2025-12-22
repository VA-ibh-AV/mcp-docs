package models

import "time"

type Usage struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	UserID       string      `json:"user_id"`
	UsageMetrics PlanMetrics `gorm:"embedded" json:"usage_metrics"`
	PeriodStart  time.Time   `json:"period_start"`
	PeriodEnd    time.Time   `json:"period_end"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
