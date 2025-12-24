package models

import "time"

// IndexingJob represents a job for indexing documents in a project
type IndexingJob struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	RequestID uint      `gorm:"not null;index" json:"request_id"`
	Url       string    `gorm:"type:varchar(255);not null" json:"url"`
	Status    string    `gorm:"not null" json:"status"` // e.g., "pending", "in_progress", "completed", "failed"
	ErrorMsg  string    `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
