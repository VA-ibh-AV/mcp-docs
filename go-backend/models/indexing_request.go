package models

import "time"

// Indexing status constants
const (
	IndexingStatusPending       = "pending"
	IndexingStatusInProgress    = "in_progress"
	IndexingStatusCrawlComplete = "crawl_complete" // Crawl done, Python agent processing
	IndexingStatusCompleted     = "completed"      // All processing done
	IndexingStatusFailed        = "failed"
)

// IndexingRequest represents a user's request to index a website
type IndexingRequest struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"not null" json:"user_id"`
	CollectionID  string    `gorm:"type:varchar(36);not null;uniqueIndex" json:"collection_id"` // UUID for LightRAG workspace isolation
	Endpoint      string    `gorm:"type:varchar(255);not null" json:"endpoint"`
	TotalJobs     int       `gorm:"not null" json:"total_jobs"`
	CompletedJobs int       `gorm:"not null" json:"completed_jobs"`
	ProjectID     uint      `gorm:"not null" json:"project_id"`
	Status        string    `gorm:"type:varchar(50);not null" json:"status"` // e.g., "pending", "in_progress", "completed", "failed"
	ErrorMsg      string    `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
