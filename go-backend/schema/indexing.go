package schema

import "time"

type CreateIndexingRequestRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Endpoint  string `json:"endpoint" binding:"required,url"`
}

type UpdateIndexingRequestStatusRequest struct {
	Status                 string `json:"status"`
	ErrorMsg               string `json:"error_msg"`
	IncrementCompletedJobs bool   `json:"increment_completed_jobs"`
}

type IndexingRequestResponse struct {
	ID            uint      `json:"id"`
	UserID        string    `json:"user_id"`
	CollectionID  string    `json:"collection_id"`
	Endpoint      string    `json:"endpoint"`
	TotalJobs     int       `json:"total_jobs"`
	CompletedJobs int       `json:"completed_jobs"`
	ProjectID     uint      `json:"project_id"`
	Status        string    `json:"status"`
	ErrorMsg      string    `json:"error_msg,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateIndexingJobRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	RequestID uint   `json:"request_id" binding:"required"`
	Url       string `json:"url" binding:"required,url"`
}

type UpdateIndexingJobStatusRequest struct {
	Status   string `json:"status" binding:"required"`
	ErrorMsg string `json:"error_msg"`
}

type IndexingJobResponse struct {
	ID        uint      `json:"id"`
	ProjectID uint      `json:"project_id"`
	RequestID uint      `json:"request_id"`
	Url       string    `json:"url"`
	Status    string    `json:"status"`
	ErrorMsg  string    `json:"error_msg,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
