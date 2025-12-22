package models

import "time"

type PlanMetrics struct {
	Projects                  int `json:"total_projects"`
	SSEExecutions             int `json:"sse_executions"`
	MaxIndexPerProject        int `json:"max_index_per_project"`
	AIPoweredSearchExecutions int `json:"ai_powered_search_executions"`
}

type Plan struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       float64     `json:"price"`
	Interval    string      `json:"interval"` // e.g., "monthly", "yearly"
	Metrics     PlanMetrics `gorm:"embedded" json:"metrics"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
