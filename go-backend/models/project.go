package models

import (
	"time"
)

type Project struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             string    `json:"user_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Url                string    `json:"url"`
	VectorDbCollection string    `json:"vector_db_collection"`
	Status             string    `json:"status"`
	IndexPageCount     int       `json:"index_page_count"`
	TotalPageCount     int       `json:"total_page_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
